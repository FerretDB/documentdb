### documentdb v1.106-0 (Unreleased) ###
- Add internal extension that provides extensions to the `rum` index. *[Feature]*

### DocumentDB v0.105.0-ferretdb-2.4.0 (July 15, 2025) ###

This version works best with FerretDB v2.4.0.

Debian and Ubuntu `.deb` packages are provided
[on the release page](https://github.com/FerretDB/documentdb/releases/tag/v0.105.0-ferretdb-2.4.0).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/deb/).

Docker images are available
[in the registry](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/docker/).
We always recommend specifying the full image tag (e.g., `17-0.105.0-ferretdb-2.4.0`, not just `17` or `17-0.105.0`)
to avoid unexpected updates.

### documentdb v0.105-0 (July 09, 2025) ###
* Support `$bucketAuto` aggregation stage, with granularity types: `POWERSOF2`, `1-2-5`, `R5`, `R10`, `R20`, `R40`, `R80`, `E6`, `E12`, `E24`, `E48`, `E96`, `E192` *[Feature]*
* Support `conectionStatus` command *[Feature]*.

### DocumentDB v0.104.0-ferretdb-2.3.0 (June 10, 2025) ###

This version works best with FerretDB v2.3.0 and v2.3.1.

> [!NOTE]
> Docker tags `XX-0.104.0-ferretdb-2.3.0` and `XX-0.104.0-ferretdb-2.3.1` point to the same images.
> Additional tags were added to accommodate FerretDB hotfix release 2.3.1.

Debian and Ubuntu `.deb` packages are provided
[on the release page](https://github.com/FerretDB/documentdb/releases/tag/v0.104.0-ferretdb-2.3.0).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/deb/).

Docker images are available
[in the registry](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/docker/).
We always recommend specifying the full image tag (e.g., `17-0.104.0-ferretdb-2.3.0`, not just `17` or `17-0.104.0`)
to avoid unexpected updates.

### documentdb v0.104-0 (June 09, 2025) ###
* Add string case support for `$toDate` operator
* Support `sort` with collation in runtime *[Feature]*
* Support collation with `$indexOfArray` aggregation operator *[Feature]*
* Support collation with arrays and objects comparisons *[Feature]*
* Support background index builds *[Bugfix]* (#36)
* Enable user CRUD by default *[Feature]*
* Enable let support for delete queries *[Feature]*. Requires `EnableVariablesSupportForWriteCommands` to be `on`.
* Enable rum_enable_index_scan as default on *[Perf]*
* Add public `documentdb-local` Docker image with gateway to GHCR
* Support `compact` command *[Feature]*. Requires `documentdb.enablecompact` GUC to be `on`.
* Enable role privileges for `usersInfo` command *[Feature]*

### DocumentDB v0.103.0-ferretdb-2.2.0 (May 09, 2025) ###

This version works best with FerretDB v2.2.0.

Debian and Ubuntu `.deb` packages are provided
[on the release page](https://github.com/FerretDB/documentdb/releases/tag/v0.103.0-ferretdb-2.2.0).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/deb/).

Docker images are available
[in the registry](https://github.com/FerretDB/documentdb/pkgs/container/postgres-documentdb).
See installation instructions [in our documentation](https://docs.ferretdb.io/installation/documentdb/docker/).
We always recommend specifying the full image tag (e.g., `17-0.103.0-ferretdb-2.2.0`, not just `17` or `17-0.103.0`)
to avoid unexpected updates.

### documentdb v0.103-0 (May 09, 2025) ###
* Support collation with aggregation and find on sharded collections *[Feature]*
* Support `$convert` on `binData` to `binData`, `string` to `binData` and `binData` to `string` (except with `format: auto`) *[Feature]*
* Fix list_databases for databases with size > 2 GB *[Bugfix]* (#119)
* Support half-precision vector indexing, vectors can have up to 4,000 dimensions *[Feature]*
* Support ARM64 architecture when building docker container *[Preview]*
* Support collation with `$documents` and `$replaceWith` stage of the aggregation pipeline *[Feature]*
* Push pg_documentdb_gw for documentdb connections *[Feature]*

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
* Support type aggregation operator `$toUUID` *[Feature]*
* Support Partial filter pushdown for `$in` predicates *[Perf]*
* Support the $dateFromString operator with full functionality *[Feature]*
* Support extended syntax for `$getField` aggregation operator (field as expression) *[Feature]*

### documentdb v0.101-0 (February 12, 2025) ###
* Push $graphlookup recursive CTE JOIN filters to index *[Perf]*
* Build pg_documentdb for PostgreSQL 17 *[Infra]* (#13)
* Enable support of currentOp aggregation stage, along with collstats, dbstats, and indexStats *[Commands]* (#52)
* Allow inlining $unwind with $lookup with `preserveNullAndEmptyArrays` *[Perf]*
* Skip loading documents if group expression is constant *[Perf]*
* Fix Merge stage not outputting to target collection *[Bugfix]* (#20)

### documentdb v0.100-0 (January 23rd, 2025) ###
Initial Release
