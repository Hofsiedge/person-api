begin;

drop function people.delete_person(uuid);

-- testing functions
do $do$
begin
    if utils.in_test_environment() then
        drop function test.test_000005_delete_person_function();
    end if;
end
$do$;

commit;
