BEGIN;

CREATE SCHEMA utils;

CREATE FUNCTION utils.in_test_environment()
RETURNS boolean
AS $sql$
    SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'test');
$sql$
LANGUAGE sql;

COMMIT;
