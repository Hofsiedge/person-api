begin;

drop function people.list_people(
    text, text, text, int, int, people.sex, char(2), real, int, int);

drop type people.people_page;

-- testing functions
do $do$
begin
    if utils.in_test_environment() then
        drop function test.test_000007_list_people_function();
    end if;
end
$do$;

commit;
