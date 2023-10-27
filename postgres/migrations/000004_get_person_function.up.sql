begin;

create function people.get_person(id uuid)
returns people.people
as $func$
declare
    result_ people.people;
begin
    if id is null then
        raise exception 'invalid person_id: %', id
            using errcode = 'invalid_parameter_value';
    end if;

    select p.* into result_
    from
        people.people p
    where
        p.person_id = id;

    if not found then
        raise exception 'person not found: %', id
            using errcode = 'no_data_found';
    end if;
    return result_;
end;
$func$
language plpgsql;


-- testing functions
do $do$
begin
    if not utils.in_test_environment() then
        return;
    end if;

    create function test.test_000003_get_person_function()
        returns setof text as $test$
        declare
            person people.people;
            id uuid;
        begin
            return next has_function('people', 'get_person', array['uuid']);

            person := (gen_random_uuid(), 'Name', 'Surname', 'Patronymic', 42, 'male', 'AA');

            insert into people.people select person.*;
            insert into people.people
                (name, surname, patronymic, age, sex, nationality)
            values
                ('N', 'S', 'P', 10, 'female', 'DE');

            return next throws_like(
                $$select people.get_person(NULL)$$,
                'invalid person_id: %',
                'throws on null id'
            );

            return next throws_like(
                format($$select people.get_person('%s')$$, gen_random_uuid()),
                'person not found: %',
                'throws on not found'
            );

            return next lives_ok(
                format($$select people.get_person('%s')$$, person.person_id),
                'can get existing person'
            );

            return next is(
                people.get_person(person.person_id),
                person,
                'returns right values'
            );
        end;
    $test$
    language plpgsql;
end
$do$;
commit;
