SET search_path TO helio_api;

#include "pg_documentdb/sql/udfs/aggregation/bson_inverse_match--0.16-0.sql"

#include "pg_documentdb/sql/udfs/metadata/empty_data_table--0.16-0.sql"
#include "pg_documentdb/sql/udfs/commands_crud/cursor_functions--0.16-0.sql"

#include "pg_documentdb/sql/udfs/geospatial/bson_gist_extensibility_functions--0.16-0.sql"
#include "udfs/aggregation/bson_geonear_functions--1.16-0.sql"
#include "pg_documentdb/sql/operators/bson_geospatial_operators--0.16-0.sql"
#include "pg_documentdb/sql/operators/bson_gist_geospatial_op_classes--0.16-0.sql"
#include "pg_documentdb/sql/operators/bson_gist_geospatial_op_classes_members--0.16-0.sql"

#include "pg_documentdb/sql/udfs/commands_crud/update--0.16-0.sql"
#include "pg_documentdb/sql/udfs/commands_crud/insert--0.16-0.sql"
#include "pg_documentdb/sql/udfs/commands_crud/delete--0.16-0.sql"

#include "pg_documentdb/sql/rbac/extension_admin_setup--0.16-0.sql"

#include "udfs/aggregation/bson_lookup_functions--1.16-0.sql"

#include "pg_documentdb/sql/udfs/commands_diagnostic/coll_stats--0.16-0.sql"

#include "pg_documentdb/sql/udfs/query/bson_dollar_selectivity--0.16-0.sql"
#include "pg_documentdb/sql/operators/bson_dollar_operators--0.16-0.sql"
#include "pg_documentdb/sql/udfs/query/bson_dollar_negation--0.16-0.sql"
#include "pg_documentdb/sql/operators/bson_dollar_negation_operators--0.16-0.sql"
#include "pg_documentdb/sql/schema/index_operator_classes_negation--0.16-0.sql"

#include "udfs/aggregation/group_aggregates_support--1.16-0.sql"
#include "udfs/aggregation/group_aggregates--1.16-0.sql"

RESET search_path;
