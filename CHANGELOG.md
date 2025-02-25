### DocumentDB v0.102.0-ferretdb-2.0.0-rc.2 (February 24, 2025) ###

This version works best with [FerretDB v2.0.0-rc.2](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.2).

Debian and Ubuntu `.deb` packages are now [provided](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.2).
Docker images are available [here](https://github.com/FerretDB/FerretDB/pkgs/container/postgres-documentdb).

### documentdb v0.102-0 (Unreleased) ###
* Support index pushdown for vector search queries *[Bugfix]*
* Support exact search for vector search queries *[Feature]*
* Inline $match with let in $lookup pipelines as JOIN Filter *[Perf]*
* Support TTL indexes *[Bugfix]* (#34)

### documentdb v0.101-0 (February 12, 2025) ###
* Push $graphlookup recursive CTE JOIN filters to index *[Perf]*
* Build pg_documentdb for PostgreSQL 17 *[Infra]* (#13)
* Enable support of currentOp aggregation stage, along with collstats, dbstats, and indexStats *[Commands]* (#52)
* Allow inlining $unwind with $lookup with `preserveNullAndEmptyArrays` *[Perf]*
* Skip loading documents if group expression is constant *[Perf]*
* Fix Merge stage not outputing to target collection *[Bugfix]* (#20)

### documentdb v0.100-0 (January 23rd, 2025) ###
Initial Release
