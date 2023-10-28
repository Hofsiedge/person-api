begin;

drop table people.people;
drop type people.sex;
drop function people.const_max_age();

drop extension pg_trgm;

drop schema people;

-- testing functions
do $do$
begin
    if utils.in_test_environment () then
        drop function test.test_000002_people_table_columns;
    end if;
end
$do$;

commit;
