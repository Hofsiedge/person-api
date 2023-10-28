begin;

create function people.update_person(
    id           uuid,
    name_        text       default null,
    surname_     text       default null, 
    patronymic_  text       default null, 
    age_         int        default null, 
    sex_         people.sex default null, 
    nationality_ char(2)    default null
)
returns void
as $func$
declare
    count_ int;
begin
    if id is null then
        raise exception 'invalid person_id: NULL'
            using errcode = 'invalid_parameter_value';
    end if;

    if (name_, surname_, patronymic_, age_, sex_, nationality_) = (null, null, null, null, null, null) then
        raise exception 'invalid arguments: nothing to update'
            using errcode = 'invalid_parameter_value';
    end if;

    update people.people p
    set
        name        = coalesce(name_,        old.name),
        surname     = coalesce(surname_,     old.surname),
        patronymic  = coalesce(patronymic_,  old.patronymic),
        age         = coalesce(age_,         old.age),
        sex         = coalesce(sex_,         old.sex),
        nationality = coalesce(nationality_, old.nationality)
    from (select * from people.people where person_id = id) old
    where
        p.person_id = id;

    get diagnostics count_ = row_count;
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

    create function test.test_000006_update_person_function()
        returns setof text as $test$
        declare
            person people.people;
            i      text;
            query  text;
            vals   text;
        begin
            return next has_function('people', 'update_person', array[
                'uuid', 'text', 'text', 'text', 'int', 'people.sex', 'char(2)'
            ]);

            person := (gen_random_uuid(), 'Name', 'Surname', 'Patronymic', 42, 'male', 'AA');

            insert into people.people select person.*;

            return next throws_like($$
                    select people.update_person(
                        gen_random_uuid(), name_ =>'Qux')
                $$,
                'person with id % not found',
                'throws on not found'
            );

            return next throws_like($$
                    select people.update_person(
                        null, 'Qux', null, null, null, null, null)
                $$,
                'invalid person_id: NULL',
                'throws on null id'
            );

            foreach vals, i in array array[
                ($$'NewName'$$,       'name'),
                ($$'NewSurname'$$,    'surname'),
                ($$'NewPatronymic'$$, 'patronymic'),
                ($$91$$,              'age'),
                ($$'female'$$,        'sex'),
                ($$'ZZ'$$,            'nationality')
            ] loop
                return next lives_ok(
                    format(
                        $$select people.update_person(%L, %s_ => %s)$$,
                        person.person_id,
                        i,
                        vals
                    ),
                    'can update just ' || i
                );
            end loop;

            return next row_eq(
                format(
                    $$select * from people.people where person_id = %L$$,
                    person.person_id
                ),
                (person.person_id, 'NewName', 'NewSurname', 
                    'NewPatronymic', 91, 'female', 'ZZ')::people.people,
                'individual updates are applied'
            );

            return next lives_ok(
                format(
                    $$select people.update_person(%L, %L, %L, %L, %L, %L, %L)$$,
                    person.person_id, person.name, person.surname, person.patronymic, 
                    person.age, person.sex, person.nationality
                ),
                'can update all fields at once'
            );

            return next row_eq(
                format(
                    $$select * from people.people where person_id = %L$$,
                    person.person_id
                ),
                person,
                'updates to all fields are applied'
            );
        end;
    $test$
    language plpgsql;
end
$do$;
commit;
