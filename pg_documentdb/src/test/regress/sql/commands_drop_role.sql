SET documentdb.next_collection_id TO 1983100;
SET documentdb.next_collection_index_id TO 1983100;

SET documentdb.maxUserLimit TO 10;
\set VERBOSITY TERSE

-- Enable role CRUD operations for testing
SET documentdb.enableRoleCrud TO ON;

-- ********* Test dropRole command basic functionality *********
-- Test dropRole of a custom role
SELECT documentdb_api.create_role('{"createRole":"customRole", "roles":["documentdb_readonly_role"]}');
SELECT documentdb_api.drop_role('{"dropRole":"customRole"}');
SELECT rolname FROM pg_roles WHERE rolname = 'customRole';

-- Test dropRole of a referenced role which will still drop regardless
SELECT documentdb_api.create_role('{"createRole":"customRole", "roles":["documentdb_readonly_role"]}');
SELECT documentdb_api.create_user('{"createUser":"userWithCustomRole", "pwd":"Valid$123Pass", "roles":[{"role":"readAnyDatabase","db":"admin"}]}');
GRANT "customRole" TO "userWithCustomRole";
SELECT documentdb_api.drop_role('{"dropRole":"customRole"}');
SELECT rolname FROM pg_roles WHERE rolname = 'customRole';
SELECT documentdb_api.drop_user('{"dropUser":"userWithCustomRole"}');

-- Test dropRole with additional fields that should be ignored
SELECT documentdb_api.create_role('{"createRole":"customRole", "roles":["documentdb_readonly_role"]}');
SELECT documentdb_api.drop_role('{"dropRole":"customRole", "lsid":"test"}');
SELECT rolname FROM pg_roles WHERE rolname = 'customRole';

-- ********* Test dropRole error inputs *********
SELECT documentdb_api.create_role('{"createRole":"customRole", "roles":["documentdb_readonly_role"]}');
SELECT rolname FROM pg_roles WHERE rolname = 'customRole';

-- Test dropRole with missing dropRole field, should fail
SELECT documentdb_api.drop_role('{}');

-- Test dropRole with empty role name, should fail
SELECT documentdb_api.drop_role('{"dropRole":""}');

-- Test dropRole with non-existent role, should fail
SELECT documentdb_api.drop_role('{"dropRole":"nonExistentRole"}');

-- Test dropRole with invalid JSON, should fail
SELECT documentdb_api.drop_role('{"dropRole":"invalidJson"');

-- Test dropRole with non-string role name, should fail
SELECT documentdb_api.drop_role('{"dropRole":1}');

-- Test dropRole with null role name, should fail
SELECT documentdb_api.drop_role('{"dropRole":null}');

-- Test dropping built-in roles, should fail
SELECT documentdb_api.drop_role('{"dropRole":"documentdb_admin_role"}');

-- Test dropRole of a system role
SELECT documentdb_api.drop_role('{"dropRole":"documentdb_bg_worker_role"}');

-- Test dropRole of non-existing role with built-in role prefix, which should fail with role not found
SELECT documentdb_api.drop_role('{"dropRole":"documentdb_role"}');

-- Test dropRole with unsupported field, should fail
SELECT documentdb_api.drop_role('{"dropRole":"customRole", "unsupportedField":"value"}');

-- Test dropRole with different casing which should fail with role not found
SELECT documentdb_api.drop_role('{"dropRole":"CUSTOMROLE"}');

-- Test dropRole when feature is disabled
SET documentdb.enableRoleCrud TO OFF;
SELECT documentdb_api.drop_role('{"dropRole":"customRole"}');
SET documentdb.enableRoleCrud TO ON;

-- Clean up and Reset settings
SELECT documentdb_api.drop_role('{"dropRole":"customRole"}');
RESET documentdb.enableRoleCrud;
