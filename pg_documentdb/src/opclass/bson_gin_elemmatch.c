/*-------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation.  All rights reserved.
 *
 * src/opclass/bson_gin_elemmatch.c
 *
 * Gin operator implementations of BSON for the $elemmatch operator.
 *
 *-------------------------------------------------------------------------
 */

#include <postgres.h>
#include <miscadmin.h>
#include <utils/builtins.h>
#include <parser/parse_coerce.h>
#include <catalog/pg_type.h>
#include <funcapi.h>

#include "opclass/helio_bson_gin_private.h"
#include "io/bson_core.h"
#include "query/bson_compare.h"
#include "operators/bson_expr_eval.h"
#include "nodes/makefuncs.h"
#include "query/query_operator.h"
#include "opclass/helio_gin_index_mgmt.h"
#include "opclass/helio_gin_index_term.h"


/* --------------------------------------------------------- */
/* Declaration of types */
/* --------------------------------------------------------- */

/*
 * For simple filters (leafs in the expression)
 * Tracks the state generated by extract_query
 * that will be dispatched for the consistent
 * function
 */
typedef struct BsonElemMatchConsistentState
{
	BsonIndexStrategy strategy;
	Pointer *extra_data;
	int32_t numKeys;
	Datum *queryKeys;
	int baseTermIndex;
	bytea *indexClassOptions;
} BsonElemMatchConsistentState;

/*
 * ElemMatch query state for nested expressions
 * where one of the nested expressions is a boolean
 * expression (AND/OR)
 * This is used in the bson_consistent check to ensure
 * that the conjunction/disjunction of nested expressions
 * is applied correctly.
 */
struct BsonElemMatchBoolFilterState;
typedef struct BsonElemMatchBoolFilterState
{
	/* whether or not this is an AND expression */
	bool isAndExpr;

	/* Any child boolean filters associated with this
	 * e.g. if there is a $and [ { $and: [ ] }]
	 */
	List *nestedBoolFilters;

	/*
	 * Any non-boolean expressions that have
	 * BsonElemMatchConsistentState
	 * associated with this boolean expression.
	 */
	List *nestedChildFilters;

	/* Pointer to the 'root' from any node. */
	struct BsonElemMatchBoolFilterState *root;
} BsonElemMatchBoolFilterState;


/*
 * Extra_data generated by elemMatch that is used by ComparePartial.
 */
typedef struct BsonElemMatchState
{
	union
	{
		/* The elemMatch filter is a query expression to be evaluated */
		BsonElemMatchIndexExprState expressionState;

		struct
		{
			/* If this is a nested filter, the index strategy of the filter */
			BsonIndexStrategy filterStrategy;

			/* If this is a nested filter, the extra_data for the term
			 * to be evaluated for ComparePartial
			 */
			Pointer filterExtraData;

			/* If this is a nested filter, a pointer to the root of
			 * the expression tree (used in Consistent) */
			BsonElemMatchBoolFilterState *filterExpressionRoot;
		};
	};

	/* Whether or not the state points to a query expression */
	bool isExpression;
} BsonElemMatchState;


/*
 * State tracked by elemMatch when the ExtractQuery
 * pulls up nested filters as a top level filter.
 * Each entry corresponds to a term generated by the
 * nested filter that can be used to query against
 * the index directly.
 */
typedef struct BsonElemMatchSimpleNestedFilterState
{
	/* The elemMatch state used during execution of the
	 * index Scan */
	BsonElemMatchState elemMatchState;

	/* The term to seek to for the strategy */
	Datum queryDatum;

	/* Whether or not this term is a partial match */
	bool partialMatch;
} BsonElemMatchSimpleNestedFilterState;


/* --------------------------------------------------------- */
/* Forward declaration */
/* --------------------------------------------------------- */

static List * ExtractSubExpressionsFromElemMatchQuery(pgbson *elemMatchQuery,
													  bytea *options);
static bool ExtractElemMatchSubExpressionWalker(List *expressions,
												BsonElemMatchBoolFilterState *state,
												List **childExpressions,
												bytea *indexOptions);
static bool ProcessFuncExprForIndexPushdown(FuncExpr *function,
											BsonElemMatchBoolFilterState *parent,
											List **elemMatchContextEntries,
											bytea *indexOptions);
static bool PathHasArrayIndexElements(const StringView *path);
static Datum * GinBsonExtractQueryElemMatchForExpression(BsonExtractQueryArgs *args);
static Datum * GinBsonExtractQueryElemMatchForNestedFilters(List *nestedExpressions,
															BsonExtractQueryArgs *args);
static bool GetElemMatchQualConsistentResult(BsonElemMatchBoolFilterState *root,
											 bool *checkArray);


/* --------------------------------------------------------- */
/* Top level exports */
/* --------------------------------------------------------- */

/*
 * Implements the ExtractQuery method for $elemMatch.
 * Walks the $elemMatch expression and builds the terms
 * necessary for pushing the $elemMatch to the index.
 */
Datum *
GinBsonExtractQueryElemMatch(BsonExtractQueryArgs *args)
{
	pgbson *query = args->query;

	/* For $elemMatch, the first step we try is to see if we can
	 * extract sub-expressions from the $elemMatch as top level
	 * expressions
	 */
	List *expressions = ExtractSubExpressionsFromElemMatchQuery(query, args->options);
	if (expressions == NIL || list_length(expressions) == 0)
	{
		/* If the nested expression cannot be evaluated against the index
		 * Construct the elemMatch based on the raw expression */
		return GinBsonExtractQueryElemMatchForExpression(args);
	}
	else
	{
		/* If the nested expression can be evaluated against the index
		 * Construct the elemMatch based on the nested filters */
		return GinBsonExtractQueryElemMatchForNestedFilters(expressions, args);
	}
}


/*
 * Implements the ComparePartial method for $elemMatch.
 * Compares the compareValue pgbson term in the index to the query given
 * the extraData for that term.
 */
int32_t
GinBsonComparePartialElemMatch(BsonIndexTerm *queryValue, BsonIndexTerm *compareValue,
							   Pointer extraData)
{
	BsonElemMatchState *elemMatchState = (BsonElemMatchState *) extraData;
	if (elemMatchState->isExpression)
	{
		return GinBsonComparePartialElemMatchExpression(queryValue, compareValue,
														&elemMatchState->expressionState);
	}
	else
	{
		/* Execute the ComparePartial based on the nested expression */
		return GinBsonComparePartialCore(elemMatchState->filterStrategy,
										 queryValue,
										 compareValue,
										 elemMatchState->filterExtraData);
	}
}


/*
 * Given an array of check booleans (one for each term queried against the index)
 * for a document, an array of extraData per term, and a number of terms,
 * validates whether that document would be a match for the $elemMatch.
 *
 */
bool
GinBsonElemMatchConsistent(bool *checkArray, Pointer *extraData, int32_t numKeys)
{
	/* First pass - check for expression or nested filters */
	bool hasExpression = false;
	for (int i = 0; i < numKeys; i++)
	{
		BsonElemMatchState *state = (BsonElemMatchState *) extraData[i];
		if (state->isExpression)
		{
			hasExpression = true;
			break;
		}
	}

	bool res = false;
	if (hasExpression)
	{
		/* if it is an elemMatch purely based on expressions,
		 * we can simply evaluate whether the expresssion was true
		 * If the expression evaluated to false, then it's still a match
		 * if the term was truncated.
		 */
		res = false;
		for (int i = 0; i < numKeys && !res; i++)
		{
			res = res || checkArray[i];
		}
	}
	else
	{
		/* Given the query state, go to the root of the query expression
		 * And evaluate the consistent state from the root
		 */
		BsonElemMatchState *state = (BsonElemMatchState *) extraData[0];
		res = GetElemMatchQualConsistentResult(state->filterExpressionRoot, checkArray);
	}

	return res;
}


/* --------------------------------------------------------- */
/* Private helper methods */
/* --------------------------------------------------------- */


/*
 * Generates the terms to be evaluated against for $elemMatch based
 * on the top level expression.
 * This generates a single term that seeks to the MinArray term
 * and scans for all arrays to find elements that match against the
 * query filter.
 */
static Datum *
GinBsonExtractQueryElemMatchForExpression(BsonExtractQueryArgs *args)
{
	pgbson *query = args->query;
	int32 *nentries = args->nentries;
	bool **partialmatch = args->partialmatch;
	Pointer **extra_data = args->extra_data;
	pgbsonelement documentElement;
	pgbson_writer bsonWriter;

	*nentries = 1;
	Datum *entries = (Datum *) palloc(sizeof(Datum) * 2);
	*partialmatch = (bool *) palloc(sizeof(bool) * 2);
	*extra_data = (Pointer *) palloc(sizeof(Pointer) * 2);

	PgbsonToSinglePgbsonElement(query, &documentElement);

	/* now create a bson for that path which has the min value for the field */
	/* we map this to an empty array since that's the smallest value we'll encounter */
	PgbsonWriterInit(&bsonWriter);
	PgbsonWriterAppendEmptyArray(&bsonWriter, documentElement.path,
								 documentElement.pathLength);
	pgbson *bson = PgbsonWriterGetPgbson(&bsonWriter);

	pgbsonelement termElement;
	PgbsonToSinglePgbsonElement(bson, &termElement);

	/* The first index term is the expression we have to evaluate */
	BsonElemMatchState *elemMatchState = palloc(sizeof(BsonElemMatchState));
	elemMatchState->expressionState.expression = GetExpressionEvalState(
		&documentElement.bsonValue,
		CurrentMemoryContext);
	elemMatchState->expressionState.isEmptyExpression = IsBsonValueEmptyDocument(
		&documentElement.bsonValue);
	elemMatchState->isExpression = true;

	Pointer *extraDataPtr = *extra_data;
	extraDataPtr[0] = (Pointer) elemMatchState;
	(*partialmatch)[0] = true;
	entries[0] = PointerGetDatum(SerializeBsonIndexTerm(&termElement,
														&args->termMetadata).
								 indexTermVal);

	if (args->termMetadata.indexTermSizeLimit > 0)
	{
		*nentries = 2;
		extraDataPtr[1] = (Pointer) elemMatchState;
		(*partialmatch)[1] = false;
		entries[1] = GenerateRootTruncatedTerm(&args->termMetadata);
	}

	return entries;
}


/*
 * Generates the terms to be evaluated against for $elemMatch based
 * on nested expressions that were pulled up to the top level.
 * This generates N terms (where N is the number of terms generated from
 * the child expressions)
 */
static Datum *
GinBsonExtractQueryElemMatchForNestedFilters(List *nestedState,
											 BsonExtractQueryArgs *args)
{
	int32 *nentries = args->nentries;
	bool **partialmatch = args->partialmatch;
	Pointer **extra_data = args->extra_data;

	int32_t entryCount = list_length(nestedState);
	*nentries = entryCount;
	*partialmatch = (bool *) palloc(sizeof(bool) * entryCount);
	*extra_data = (Pointer *) palloc(sizeof(Pointer) * entryCount);
	Datum *results = (Datum *) palloc(sizeof(Datum) * entryCount);

	ListCell *stateCell;
	int index = 0;
	foreach(stateCell, nestedState)
	{
		BsonElemMatchSimpleNestedFilterState *queryState =
			(BsonElemMatchSimpleNestedFilterState *) lfirst(stateCell);
		results[index] = queryState->queryDatum;
		(*partialmatch)[index] = queryState->partialMatch;

		/* Store the partial match extra_data query state for elemMatch */
		(*extra_data)[index] = (Pointer) & queryState->elemMatchState;
		index++;
	}

	return results;
}


/*
 * Inline convenience method that creates a BsonElemMatchBoolFilterState.
 * This is used when walking the query expression and generating bool expressions.
 */
inline static BsonElemMatchBoolFilterState *
MakeElemMatchBoolExprState(bool isAndExpr, BsonElemMatchBoolFilterState *root)
{
	BsonElemMatchBoolFilterState *state = (BsonElemMatchBoolFilterState *) palloc(
		sizeof(BsonElemMatchBoolFilterState));
	state->isAndExpr = isAndExpr;
	state->nestedBoolFilters = NIL;
	state->nestedChildFilters = NIL;
	state->root = root;
	return state;
}


/*
 * Walks the $elemMatch expression for child expressions and produces
 * a list of clauses (implicit AND) that are equivalent to the
 * elemMatch that can produce the same results as running the
 * expression directly against the index. Note that this can still have
 * false positives, but should not return false negatives.
 *
 * If it is not possible to return a clause that produces false negatives
 * returns NIL.
 */
static List *
ExtractSubExpressionsFromElemMatchQuery(pgbson *elemMatchQuery, bytea *options)
{
	pgbsonelement singleElement;
	PgbsonToSinglePgbsonElement(elemMatchQuery, &singleElement);

	if (IsBsonValueEmptyDocument(&singleElement.bsonValue))
	{
		/* Empty query defaults to expression evaluation */
		return NIL;
	}

	/* Convert the pgbson query into a query AST that processes bson */
	Expr *expr = CreateQualForBsonExpression(&singleElement.bsonValue,
											 singleElement.path);

	/* Get the underlying list of expressions that are AND-ed */
	List *clauses = make_ands_implicit(expr);

	/* We create a root expression to track AND/OR expressions */
	bool isAndExpr = true;
	BsonElemMatchBoolFilterState *root = MakeElemMatchBoolExprState(isAndExpr, NULL);

	/* Now walk the clauses and build the List of child expressions that can be
	 * pushed to the index */
	root->root = root;
	List *childExpressions = NIL;
	if (ExtractElemMatchSubExpressionWalker(clauses, root, &childExpressions, options))
	{
		return childExpressions;
	}
	else
	{
		return NIL;
	}
}


/*
 * Walks the list of clauses and converts them into expressions that can
 * be pushed down to the index that are functionally equivalent to the
 * top level $elemmatch expression. If the expressions cannot be safely
 * converted to sub-expressions and could return false negatives, then
 * returns false. else returns true and adds the expressions to the
 * List of child expressions.
 */
static bool
ExtractElemMatchSubExpressionWalker(List *clauses, BsonElemMatchBoolFilterState *parent,
									List **childExpressions, bytea *indexOptions)
{
	CHECK_FOR_INTERRUPTS();
	check_stack_depth();

	ListCell *clauseCell;
	foreach(clauseCell, clauses)
	{
		Expr *clause = (Expr *) lfirst(clauseCell);
		if (IsA(clause, FuncExpr))
		{
			FuncExpr *funcExpr = (FuncExpr *) clause;
			if (!ProcessFuncExprForIndexPushdown(funcExpr, parent, childExpressions,
												 indexOptions))
			{
				return NIL;
			}
		}
		else if (IsA(clause, BoolExpr))
		{
			BoolExpr *expr = (BoolExpr *) clause;

			if (expr->boolop == NOT_EXPR)
			{
				/* Not is a negation operator and can produce false negatives.
				 * Cannot safely push this expression as-is.
				 */
				return NIL;
			}

			bool isAndExpr = expr->boolop == AND_EXPR;
			BsonElemMatchBoolFilterState *state = MakeElemMatchBoolExprState(isAndExpr,
																			 parent->root);

			/* Now walk the child expressions for the BOOL expression */
			if (ExtractElemMatchSubExpressionWalker(expr->args, state, childExpressions,
													indexOptions))
			{
				parent->nestedBoolFilters = lappend(parent->nestedBoolFilters, state);
				return true;
			}
			else
			{
				return false;
			}
		}
		else
		{
			/* If there is an expression that is neither a FuncExpr
			 * nor a BoolExpr, we don't know how to safely push this to
			 * the index. return NIL.
			 */
			return NIL;
		}
	}

	return true;
}


/*
 * Takes a single Mongo comparison FuncExpr and converts it to an OpExpr
 * That can be evaluated against the index if it can be safely pushed down
 * as a functionally equivalent clause to the $elemMatch that will not produce
 * false negatives. Valid terms will be appended to the list of
 * terms. Returns true if this expression is valid. Otherwise returns false.
 */
static bool
ProcessFuncExprForIndexPushdown(FuncExpr *function,
								BsonElemMatchBoolFilterState *parent,
								List **elemMatchContextEntries,
								bytea *indexOptions)
{
	Expr *secondArg = lsecond(function->args);
	Assert(IsA(secondArg, Const));
	Const *constVal = (Const *) secondArg;
	Assert(!constVal->constisnull);
	pgbsonelement valueElement;
	pgbson *bsonVal = DatumGetPgBson(constVal->constvalue);
	PgbsonToSinglePgbsonElement(bsonVal, &valueElement);

	StringView stringView = {
		.string = valueElement.path, .length = valueElement.pathLength
	};

	/* If the path has an array index, we can be traversing into nested arrays and yield
	 * false negatives - use expression evaluation in this case */
	if (PathHasArrayIndexElements(&stringView))
	{
		return false;
	}

	/* Lookup the func in the set of operators */
	const MongoIndexOperatorInfo *operator =
		GetMongoIndexOperatorInfoByPostgresFuncId(
			function->funcid);


	/* If this path generates a term that is incompatible with this index, bail */
	if (!ValidateIndexForQualifierValue(indexOptions, constVal->constvalue,
										operator->indexStrategy))
	{
		return false;
	}

	switch (operator->indexStrategy)
	{
		case BSON_INDEX_STRATEGY_INVALID:
		case BSON_INDEX_STRATEGY_DOLLAR_ELEMMATCH:
		case BSON_INDEX_STRATEGY_DOLLAR_NOT_IN:
		case BSON_INDEX_STRATEGY_DOLLAR_NOT_EQUAL:
		case BSON_INDEX_STRATEGY_DOLLAR_EXISTS:
		case BSON_INDEX_STRATEGY_DOLLAR_NOT_GT:
		case BSON_INDEX_STRATEGY_DOLLAR_NOT_GTE:
		case BSON_INDEX_STRATEGY_DOLLAR_NOT_LT:
		case BSON_INDEX_STRATEGY_DOLLAR_NOT_LTE:
		{
			/* These operators produce incorrect results unless evaluated as an expression */

			/* Specifically, negation operators can produce false negatives.
			 * This is similar to $exists: false as well.
			 * For $exists: true scans all paths, it's cheaper to just let
			 * elemMatch expression handle it (fewer terms scanned)
			 */
			return false;
		}

		case BSON_INDEX_STRATEGY_DOLLAR_EQUAL:
		{
			/* Special case for $eq, if it's null it can produce false negatives. */
			if (valueElement.bsonValue.value_type == BSON_TYPE_NULL)
			{
				return NULL;
			}
			else
			{
				break;
			}
		}

		case BSON_INDEX_STRATEGY_DOLLAR_IN:
		{
			/* Special case for $in, if it's null it can produce false negatives. */
			bson_iter_t arrayIterator;
			if (valueElement.bsonValue.value_type == BSON_TYPE_ARRAY &&
				bson_iter_init_from_data(&arrayIterator,
										 valueElement.bsonValue.value.v_doc.data,
										 valueElement.bsonValue.value.v_doc.
										 data_len))
			{
				while (bson_iter_next(&arrayIterator))
				{
					if (bson_iter_type(&arrayIterator) == BSON_TYPE_NULL)
					{
						return NULL;
					}
				}
			}

			break;
		}

		case BSON_INDEX_STRATEGY_DOLLAR_ALL:
		{
			/* $all evaluates nested arrays. This can produce false negatives
			 * against the index since we can't evaluate nested arrays.
			 * return NULL
			 */
			return NULL;
		}

		default:
		{
			break;
		}
	}

	/* Call ExtractQuery on the underlying operator. This will generate term datums
	 * and extra data. Pull this up into the state so that we can execute it for
	 * elemMatch.
	 */
	bool *partialMatch = NULL;

	BsonElemMatchConsistentState *consistentState = palloc0(
		sizeof(BsonElemMatchConsistentState));
	consistentState->strategy = operator->indexStrategy;
	consistentState->baseTermIndex = list_length(*elemMatchContextEntries);
	consistentState->indexClassOptions = indexOptions;
	BsonExtractQueryArgs args =
	{
		.query = bsonVal,
		.nentries = &consistentState->numKeys,
		.partialmatch = &partialMatch,
		.extra_data = &consistentState->extra_data,
		.options = indexOptions,
		.termMetadata = GetIndexTermMetadata(indexOptions)
	};

	Datum *entries = GinBsonExtractQueryCore(operator->indexStrategy, &args);
	consistentState->queryKeys = entries;

	for (int i = 0; i < consistentState->numKeys; i++)
	{
		BsonElemMatchSimpleNestedFilterState *state = palloc0(
			sizeof(BsonElemMatchSimpleNestedFilterState));
		state->elemMatchState.isExpression = false;
		state->elemMatchState.filterStrategy = operator->indexStrategy;
		state->partialMatch = false;
		state->partialMatch = partialMatch != NULL && partialMatch[i];
		state->elemMatchState.filterExtraData = NULL;
		if (consistentState->extra_data != NULL)
		{
			state->elemMatchState.filterExtraData = consistentState->extra_data[i];
		}

		state->queryDatum = entries[i];
		state->elemMatchState.filterExpressionRoot = parent->root;

		*elemMatchContextEntries = lappend(*elemMatchContextEntries, state);
		parent->nestedChildFilters = lappend(parent->nestedChildFilters, consistentState);
	}

	return true;
}


/*
 * Walks the query tree of AND/OR and child expressions to evaluate whether
 * a document matches the filter terms based on the checkArray.
 */
static bool
GetElemMatchQualConsistentResult(BsonElemMatchBoolFilterState *expression,
								 bool *checkArray)
{
	ListCell *cell;
	if (expression->isAndExpr)
	{
		/* Check that child expressions match - if any don't match, it's a mismatch for AND */
		foreach(cell, expression->nestedChildFilters)
		{
			BsonElemMatchConsistentState *state = lfirst(cell);
			bool recheckIgnore;
			bool res = GinBsonConsistentCore(state->strategy,
											 &checkArray[state->baseTermIndex],
											 state->extra_data,
											 state->numKeys,
											 &recheckIgnore,
											 state->queryKeys,
											 state->indexClassOptions);
			if (!res)
			{
				return false;
			}
		}

		/* Check that boolean expressions match - if any don't match, it's a mismatch for AND */
		foreach(cell, expression->nestedBoolFilters)
		{
			BsonElemMatchBoolFilterState *state = lfirst(cell);
			if (!GetElemMatchQualConsistentResult(state, checkArray))
			{
				return false;
			}
		}

		/* All expressions match - it's a match for AND */
		return true;
	}
	else
	{
		/* Check that child expressions match - if any match, it's a match for OR */
		foreach(cell, expression->nestedChildFilters)
		{
			BsonElemMatchConsistentState *state = lfirst(cell);
			bool recheckIgnore;
			bool res = GinBsonConsistentCore(state->strategy,
											 &checkArray[state->baseTermIndex],
											 state->extra_data,
											 state->numKeys,
											 &recheckIgnore,
											 state->queryKeys,
											 state->indexClassOptions);
			if (res)
			{
				return true;
			}
		}

		/* Check that child bool expressions match - if any match, it's a match for OR */
		foreach(cell, expression->nestedBoolFilters)
		{
			BsonElemMatchBoolFilterState *state = lfirst(cell);
			if (GetElemMatchQualConsistentResult(state, checkArray))
			{
				return true;
			}
		}

		/* Nothing matched - it's a mismatch for OR */
		return false;
	}
}


/*
 * Walks a dotted path specified by 'a.b.c.d' and checks if any
 * component is an array index (or deals with array index paths).
 * This is defined as any path that can be an integer type.
 */
static bool
PathHasArrayIndexElements(const StringView *path)
{
	StringView subPath = StringViewFindPrefix(path, '.');
	if (subPath.length == 0)
	{
		/* not a dotted path - top level field */
		return false;
	}

	StringView newPath = *path;

	do {
		if (isdigit(subPath.string[0]))
		{
			char *endptr;
			strtol(subPath.string, &endptr, 10);
			if (*endptr == '\0' || *endptr == '.')
			{
				/* integer path */
				return true;
			}
		}

		newPath = StringViewSubstring(&newPath, subPath.length + 1);
		subPath = StringViewFindPrefix(&newPath, '.');
	} while (subPath.length > 0);

	if (newPath.length > 0)
	{
		if (isdigit(newPath.string[0]))
		{
			char *endptr;
			strtol(newPath.string, &endptr, 10);
			if (*endptr == '\0')
			{
				/* integer path */
				return true;
			}
		}
	}

	return false;
}
