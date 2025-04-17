### documentdb v0.103-0 (Unreleased) ###
* Support collation with aggregation and find on sharded collections *[Feature]*
* Support `$convert` on `binData` to `binData`, `string` to `binData` and `binData` to `string` (except with `format: auto`) *[Feature]*
* Fix list_databases for databases with size > 2 GB *[Bugfix]* (#119)
* Support ARM64 architecture when building docker container *[Preview]*
* Support collation with `$documents` and `$replceWith` stage of the aggregation pipeline *[Feature]*

### DocumentDB v0.102.0-ferretdb-2.1.0 (April 2, 2025) ###

> [!CAUTION]
> Please note that due to incompatibilities in our previous releases, they can't be updated in place,
> even with a manual `ALTER EXTENSION UPDATE` query or other means.
> A new clean installation into an empty data directory/volume is required.
> All data should be backed up with `mongodump`/`mongoexport` before
> and restored with `mongorestore`/`mongoimport` after.
>
> We expect future updates to be much smoother.

This version works best with the upcoming FerretDB v2.1.0.

Debian and Ubuntu `.deb` packages are provided [on the release page](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.1.0).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/deb/).

Docker images are available [in the registry](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/docker/).
We always recommend specifying the full image tag (e.g., `17-0.102.0-ferretdb-2.1.0`, not just `17` or `17-0.102.0`) to avoid unexpected updates.

### DocumentDB v0.102.0-ferretdb-2.0.0 (GA) (March 5, 2025) ###

This version works best with [FerretDB v2.0.0 (GA)](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0).

Debian and Ubuntu `.deb` packages are provided [on the release page](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/deb/).

Docker images are available [in the registry](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/docker/).
We always recommend specifying the full image tag (e.g., `17-0.102.0-ferretdb-2.0.0`, not just `17` or `17-0.102.0`) to avoid unexpected updates.

### documentdb v0.102-0 (March 26, 2025) ###
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
* Support collation with `$setEquals`, `$setUnion`, `$setIntersection`, `$setDifference`, `$setIsSubset` in the aggregation pipeline (Opt-in) *[Feature]*
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
