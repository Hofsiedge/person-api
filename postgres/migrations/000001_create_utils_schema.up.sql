begin;

create schema utils;

create function utils.in_test_environment()
returns boolean
as $sql$
    select exists (select 1 from pg_namespace where nspname = 'test');
$sql$
language sql;

commit;
