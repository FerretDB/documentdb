SET search_path TO documentdb_api,documentdb_core,documentdb_api_catalog;

SET citus.next_shard_id TO 5880000;
SET documentdb.next_collection_id TO 588000;
SET documentdb.next_collection_index_id TO 588000;

-- if documentdb_extended_rum exists, set alternate index handler
SELECT pg_catalog.set_config('documentdb.alternate_index_handler_name', 'extended_rum', false), extname FROM pg_extension WHERE extname = 'documentdb_extended_rum';

SELECT documentdb_api_internal.create_indexes_non_concurrently('prep_unique_db', '{ "createIndexes": "collection", "indexes": [ { "name": "a_1", "key": { "a": 1 }, "storageEngine": { "enableOrderedIndex": true, "buildAsUnique": true } } ] }', TRUE);

\d documentdb_data.documents_588001;

-- inserting and querying works fine.
select COUNT(documentdb_api.insert_one('prep_unique_db', 'collection', FORMAT('{ "a": %s, "b": %s }', i, 100-i)::bson)) FROM generate_series(1, 100) i;

-- insert a duplicate, should not fail
SELECT documentdb_api.insert_one('prep_unique_db', 'collection', '{ "a": 1 }');

ANALYZE documentdb_data.documents_588001;

set citus.propagate_set_commands to 'local';

BEGIN;
set local documentdb.enableExtendedExplainPlans to on;
set local enable_seqscan to off;
EXPLAIN (ANALYZE ON, VERBOSE ON, COSTS OFF, TIMING OFF, SUMMARY OFF) select document FROM bson_aggregation_find('prep_unique_db', '{ "find": "collection", "filter": {"a": {"$gt": 2}} }');
ROLLBACK;

BEGIN;
set local documentdb.enableExtendedExplainPlans to on;
set local documentdb.enableIndexOrderByPushdown to on;
EXPLAIN (ANALYZE ON, VERBOSE ON, COSTS OFF, TIMING OFF, SUMMARY OFF) select document FROM bson_aggregation_find('prep_unique_db', '{ "find": "collection", "filter": {"a": {"$gt": 2}}, "sort": {"a": 1} }');
ROLLBACK;


-- multiple fields work fine
SELECT documentdb_api_internal.create_indexes_non_concurrently('prep_unique_db', '{ "createIndexes": "collection", "indexes": [ { "name": "a_1_b_1", "key": { "a": 1, "b": 1 }, "storageEngine": { "enableOrderedIndex": true, "buildAsUnique": true } } ] }', TRUE);

SELECT bson_dollar_unwind(cursorpage, '$cursor.firstBatch') FROM documentdb_api.list_indexes_cursor_first_page('prep_unique_db', '{ "listIndexes": "collection" }') ORDER BY 1;

\d documentdb_data.documents_588001;
ANALYZE documentdb_data.documents_588001;

BEGIN;
set local documentdb.enableExtendedExplainPlans to on;
set local enable_seqscan to off;
EXPLAIN (ANALYZE ON, VERBOSE ON, COSTS OFF, TIMING OFF, SUMMARY OFF) select document FROM bson_aggregation_find('prep_unique_db', '{ "find": "collection", "filter": {"a": {"$gt": 2}, "b": 3} }');
ROLLBACK;

-- test error paths
SELECT documentdb_api_internal.create_indexes_non_concurrently('prep_unique_db', '{"createIndexes": "collection", "indexes": [{"key": {"a": 1 }, "storageEngine": { "enableOrderedIndex": false, "buildAsUnique": true }, "name": "invalid"}]}',true);
SELECT documentdb_api_internal.create_indexes_non_concurrently('prep_unique_db', '{"createIndexes": "collection", "indexes": [{"key": {"b": 1 }, "unique": true, "storageEngine": { "enableOrderedIndex": true, "buildAsUnique": true }, "name": "invalid"}]}',true);
