begin;

create function people.create_person(
    name_        text,
    surname_     text, 
    patronymic_  text, 
    age_         int, 
    sex_         people.sex, 
    nationality_ char(2)
)
returns uuid
as $sql$
    insert into people.people 
        (name, surname, patronymic, age, sex, nationality)
    values
        (name_, surname_, patronymic_, age_, sex_, nationality_)
    returning
        person_id;
$sql$
language sql;


-- testing functions
do $do$
begin
    if not utils.in_test_environment() then
        return;
    end if;

    create function test.test_000004_create_person_function()
        returns setof text as $test$
        declare
            person people.people;
            id uuid;
        begin
            return next has_function('people', 'create_person',
                array['text', 'text', 'text', 'int', 'people.sex', 'char(2)']);

            person := (gen_random_uuid(), 'Name', 'Surname', 'Patronymic', 42, 'male', 'AA');

            return next lives_ok(
                format($$select people.create_person(%L, %L, %L, %L, %L, %L)$$,
                    person.name, person.surname, person.patronymic,
                    person.age, person.sex, person.nationality
                ),
                'can create valid person'
            );

            person.person_id = people.create_person(
                person.name, person.surname, person.patronymic,
                person.age, person.sex, person.nationality
            );
            return next row_eq(
                format(
                    $$select * from people.people where person_id = %L$$,
                    person.person_id
                ),
                person,
                'uses provided values'
            );
        end;
    $test$
    language plpgsql;
end
$do$;
commit;
