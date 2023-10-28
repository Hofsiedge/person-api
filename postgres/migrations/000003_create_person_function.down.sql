begin;

drop function people.create_person(text, text, text, int, people.sex, char(2));

-- testing functions
do $do$
begin
    if utils.in_test_environment() then
        drop function test.test_000004_create_person_function();
    end if;
end
$do$;

commit;
