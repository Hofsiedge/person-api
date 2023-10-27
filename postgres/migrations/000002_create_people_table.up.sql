begin;

create schema people;

create function people.const_max_age()
returns int
as $$
    select 125
$$ language sql immutable;

create type people.sex as enum (
    'male',
    'female'
);

create table people.people (
    person_id   uuid             primary key default (gen_random_uuid()),
    name        text             not null,
    surname     text             not null,
    patronymic  text             not null,
    age         int              not null,
    sex         people.sex       not null,
    nationality char(2)          not null,
    constraint valid_name        check (name <> ''),
    constraint valid_surname     check (surname <> ''),
    constraint valid_age         check (age between 0 and people.const_max_age()),
    constraint valid_nationality check (nationality ~ '[A-Z][A-Z]')
);

create index people_by_age on people.people (age);
create index people_by_nationality on people.people (nationality);

-- load the trigram extension to use for word similarity indexes
create extension pg_trgm;

create index people_by_name on people.people
    using gist (name gist_trgm_ops);

create index people_by_surname on people.people
    using gist (surname gist_trgm_ops);

create index people_by_patronymic on people.people
    using gist (patronymic gist_trgm_ops);

-- testing functions
do $do$
begin
    if not utils.in_test_environment() then
        return;
    end if;

    create function test.test_000002_people_table_columns()
        returns setof text as $test$
        declare
            cols text[] = array[
                'person_id', 'name', 'surname', 'patronymic',
                'age', 'sex', 'nationality'
            ];
            i text;
        begin
            return next has_schema('people');
            return next tables_are('people', array['people']);
            return next columns_are('people', 'people', cols);

            -- check column types
            for i in (
                select col_type_is('people', 'people', col, typ, 
                        format('people.people.%s is of type %s', col, typ))
                from (
                values 
                    ('person_id',   'uuid'), 
                    ('name',        'text'),
                    ('surname',     'text'),
                    ('patronymic',  'text'),
                    ('age',         'int'),
                    ('sex',         'people.sex'),
                    ('nationality', 'char(2)')
                ) as t(col, typ)
            ) loop
                return next i;
            end loop;

            -- check not null constraint
            foreach i in array cols loop
                return next(
                    select col_not_null('people', 'people', i, 
                        'people.people.' || i || ' is not null')
                );
            end loop;

            -- check other constraints
             return next col_has_check('people', 'people', 'age',
                'todo.tasks.age must be checked');

             return next col_has_check('people', 'people', 'nationality',
                'todo.tasks.nationality must be checked');

            -- check indexes
            foreach i in array array[
                'person_id', 'name', 'surname',
                'patronymic', 'age', 'nationality'
            ] loop
                return next is_indexed('people', 'people', array[i]);
            end loop;

            -- check valid data
            return next lives_ok(
                $$insert into people.people (name, surname, patronymic, age, sex, nationality)
                    values 
                        ('Pyotr', 'Ivanov', 'Semenovich', 42, 'male', 'RU'), 
                        ('John', 'Smith', '', 30, 'male', 'US'),
                        ('Klara', 'Hummel', '', 15, 'female', 'DE')
                $$,
                'can create a valid person'
            );

            -- check invalid name
            return next throws_like($$
                    insert into people.people
                        (name, surname, patronymic, age, sex, nationality)
                    values 
                        ('', 'Ivanov', 'Semenovich', 10, 'male', 'RU')
                $$,
                '%violates check constraint "valid_name"',
                'can''t use empty name'
            );

            -- check invalid surname
            return next throws_like($$
                    insert into people.people
                        (name, surname, patronymic, age, sex, nationality)
                    values 
                        ('Pyotr', '', 'Semenovich', 10, 'male', 'RU')
                $$,
                '%violates check constraint "valid_surname"',
                'can''t use empty surname'
            );

            -- check invalid age
            foreach i in array array['-10', (people.const_max_age() + 10)::text] loop
                return next throws_like(
                    format($$
                        insert into people.people
                            (name, surname, patronymic, age, sex, nationality)
                        values 
                            ('Pyotr', 'Ivanov', 'Semenovich', %s, 'male', 'RU')
                        $$,
                        i
                    ),
                    '%violates check constraint "valid_age"',
                    format('can''t use invalid age (%s)', i)
                );
            end loop;

            -- check invalid nationality
            foreach i in array array['A1', 'ru', 'uS', 'Gb', 'X.'] loop
                return next throws_like(
                    format($$
                        insert into people.people
                            (name, surname, patronymic, age, sex, nationality)
                        values 
                            ('Pyotr', 'Ivanov', 'Semenovich', 50, 'male', '%s')
                        $$,
                        i
                    ),
                    '%violates check constraint "valid_nationality"',
                    format('can''t use invalid nationality (%s)', i)
                );
            end loop;
        end;
    $test$
    language plpgsql;
end
$do$;
commit;
