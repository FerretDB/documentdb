### DocumentDB v0.102.0-ferretdb-2.0.0 (GA) (March 5, 2025) ###

This version works best with [FerretDB v2.0.0 (GA)](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0).

Debian and Ubuntu `.deb` packages are provided [here](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0).
Docker images are available [here](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).
We always recommend specifying the full version (e.g., `:17-0.102.0-ferretdb-2.0.0`) to avoid unexpected updates.

### DocumentDB v0.102.0-ferretdb-2.0.0-rc.5 (March 4, 2025) ###

This version works best with [FerretDB v2.0.0-rc.5](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.5).

Debian and Ubuntu `.deb` packages are provided [here](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.5).
Docker images are available [here](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).

Please note that naming schemes for both `.deb` packages and Docker image tags have been changed.
We always recommend specifying the full version (e.g., `:17-0.102.0-ferretdb-2.0.0-rc.5`) to avoid unexpected updates.

### DocumentDB v0.102.0-ferretdb-2.0.0-rc.2 (February 24, 2025) ###

This version works best with [FerretDB v2.0.0-rc.2](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.2).

Debian and Ubuntu `.deb` packages are now [provided](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.2).
Docker images are available [here](https://github.com/FerretDB/FerretDB/pkgs/container/postgres-documentdb).

### documentdb v0.102-0 (Unreleased) ###
* Support index pushdown for vector search queries *[Bugfix]*
* Support exact search for vector search queries *[Feature]*
* Inline $match with let in $lookup pipelines as JOIN Filter *[Perf]*
* Support TTL indexes *[Bugfix]* (#34)
* Support joining between postgres and documentdb tables *[Feature]* (#61)
* Support current_op command *[Feature]* (#59)
* Support for list_databases command *[Feature]* (#45)
* Disable analyze statistics for unique index uuid columns which improves resource usage *[Perf]*
* Support collation with `$expr`, `$in`, `$cmp`, `$eq`, `$ne`, `$lt`, `$lte`, `$gt`, `$gte` comparison operators (Opt-in) *[Feature]*
* Support collation in `find`, aggregation `$project`, `$redact`, `$set`, `$addFields`, `$replaceRoot` stages (Opt-in) *[Feature]*
* Support collation with `$setEquals`, `$setUnion`, `$setIntersection`, `$setDifference`, `$setIsSubet` in the aggregation pipeline (Opt-in) *[Feature]*
* Support unique index truncation by default with new operator class *[Feature]*
* Top level aggregate command `let` variables support for `$geoNear` stage *[Feature]*
* Enable Backend Command support for Statement Timeout *[Feature]*
* Support type aggregation operator `$toUUID`. *[Feature]* 
* Support Partial filter pushdown for `$in` predicates *[Perf]*
* Support the $dateFromString operator with full functionality *[Feature]*
* Support extended syntax for `$getField` aggregation operator. Now the value of 'field' could be an expression that resolves to a string. *[Feature]*

### documentdb v0.101-0 (February 12, 2025) ###
* Push $graphlookup recursive CTE JOIN filters to index *[Perf]*
* Build pg_documentdb for PostgreSQL 17 *[Infra]* (#13)
* Enable support of currentOp aggregation stage, along with collstats, dbstats, and indexStats *[Commands]* (#52)
* Allow inlining $unwind with $lookup with `preserveNullAndEmptyArrays` *[Perf]*
* Skip loading documents if group expression is constant *[Perf]*
* Fix Merge stage not outputing to target collection *[Bugfix]* (#20)

### documentdb v0.100-0 (January 23rd, 2025) ###
Initial Release
