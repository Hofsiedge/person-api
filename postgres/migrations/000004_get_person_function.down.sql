begin;

drop function people.get_person(uuid);

-- testing functions
do $do$
begin
    if utils.in_test_environment() then
        drop function test.test_000003_get_person_function();
    end if;
end
$do$;

commit;
