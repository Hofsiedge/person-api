begin;

drop function people.update_person(
    uuid, text, text, text, int, people.sex, char(2));

-- testing functions
do $do$
begin
    if utils.in_test_environment() then
        drop function test.test_000006_update_person_function();
    end if;
end
$do$;

commit;
