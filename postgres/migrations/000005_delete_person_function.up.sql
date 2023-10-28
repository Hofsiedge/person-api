begin;

create function people.delete_person(id uuid)
returns void
as $func$
declare
    count_ int;
begin
    if id is null then
        raise exception 'invalid person_id'
            using errcode = 'invalid_parameter_value';
    end if;

    delete from people.people where person_id = id;

    get diagnostics count_ = ROW_COUNT;
    if count_ = 0 then
        raise exception 'person with id % not found', id
            using errcode = 'no_data_found';
    end if;
end;
$func$
language plpgsql;


-- testing functions
do $do$
begin
    if not utils.in_test_environment() then
        return;
    end if;

    create function test.test_000005_delete_person_function()
        returns setof text as $test$
        declare
            person people.people;
        begin
            return next has_function('people', 'delete_person', array['uuid']);

            return next throws_ok(
                $$select people.delete_person(NULL)$$,
                'invalid person_id',
                'throws on null id'
            );

            return next throws_like(
                $$select people.delete_person(gen_random_uuid())$$,
                'person with id % not found',
                'throws on not found'
            );

            person := (gen_random_uuid(), 'Name', 'Surname', 'Patronymic', 42, 'male', 'AA');

            insert into people.people select person.*;
            insert into people.people
                (name, surname, patronymic, age, sex, nationality)
            values (
                person.name, person.surname, person.patronymic,
                person.age, person.sex, person.nationality
            );

            return next lives_ok(
                format($$select people.delete_person(%L)$$, person.person_id),
                'can delete existing person'
            );

            return next is(
                (exists (select * from people.people 
                    where person_id = person.person_id)),
                false,
                'deletes the person with specified id'
            );

            return next (
                select ok(
                    count(*) = 1,
                    'does not delete other records'
                ) from people.people
            );
        end;
    $test$
    language plpgsql;
end
$do$;
commit;
