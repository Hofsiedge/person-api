begin;

create type people.people_page as (
    people people.people[],
    current_offset int,
    current_limit int,
    total int
);

create function people.list_people(
    name_        text       default null,
    surname_     text       default null,
    patronymic_  text       default null,
    age_min      int        default null,
    age_max      int        default null,
    sex_         people.sex default null,
    nationality_ char(2)    default null,
    threshold    real       default 0,
    offset_      int        default 0,
    limit_       int        default null
)
returns people.people_page
as $func$
declare
    count_ int;
    page   people.people_page;
begin
    with matched_people as (
        select 
            p.person_id,
            p.name,
            p.surname,
            p.patronymic,
            p.age,
            p.sex,
            p.nationality,
            (
                (case 
                    when name_ is null then 0.0
                    else word_similarity(name_, p.name)       
                end)
                +
                (case
                    when surname_ is null then 0.0
                    else word_similarity(surname_, p.surname)
                end)
                +
                (case
                    when patronymic_ is null
                        then 0
                    when (patronymic_ = '' or p.patronymic = '')
                        then (patronymic_ = p.patronymic)::int
                    else word_similarity(patronymic_, p.patronymic)
                end)
            ) / 3.0 as total_similarity
        from people.people p
        where
            ((age_min      is null) or (age_min      <= p.age))         and
            ((age_max      is null) or (age_max      >= p.age))         and
            ((nationality_ is null) or (nationality_ =  p.nationality)) and
            ((sex_         is null) or (sex_         =  p.sex))         and
            ((patronymic_  is null) or ((patronymic_ = '') = (p.patronymic = '')))
    )
    select into page.people, page.total
        array(
            select (
                m.person_id, m.name, m.surname,
                m.patronymic, m.age, m.sex, m.nationality
            )::people.people
            from matched_people m
            where m.total_similarity >= threshold
            order by 
                m.total_similarity desc,
                m.surname          asc,
                m.name             asc,
                m.patronymic       asc,
                m.age              asc,
                m.sex              asc,
                m.nationality      asc
            offset offset_
            limit limit_
        ),
        count(*)
    from matched_people;

    page.current_offset := offset_;
    page.current_limit := limit_;

    return page;
end;
$func$
language plpgsql;


-- testing functions
do $do$
begin
    if not utils.in_test_environment() then
        return;
    end if;

    create function test.test_000007_list_people_function()
        returns setof text as $test$
        declare
            -- data is sorted just like the function result for convenience
            vals people.people[] := array[
                (gen_random_uuid(), 'Alexander', 'Ivanov',   'Alexeyevich', 28, 'male',   'RU'),
                (gen_random_uuid(), 'Alexandra', 'Ivanova',  'Alexeyevna',  23, 'female', 'RU'),
                (gen_random_uuid(), 'Peter',     'Jackson',  '',            30, 'male',   'US'),
                (gen_random_uuid(), 'Ivan',      'Semyonov', 'Petrovich',   10, 'male',   'RU'),
                (gen_random_uuid(), 'Alexander', 'Sergeyev', 'Alexeyevich', 50, 'male',   'RU')
            ];
            val people.people;
        begin
            -- fill with data
            foreach val in array vals loop
                insert into people.people select val.*;
            end loop;

            return next lives_ok(
                $$select people.list_people()$$,
                'can call with null arguments'
            );

            return next is(
                people.list_people(),
                (vals, 0, null, 5)::people.people_page,
                'returns all records on null arguments' 
            );

            return next is(
                people.list_people(nationality_ => 'RU'),
                (vals[:2] || vals[4:], 0, null, 4)::people.people_page,
                'can filter by nationality' 
            );

            return next is(
                people.list_people(sex_ => 'female'),
                (vals[2:2], 0, null, 1)::people.people_page,
                'can filter by sex' 
            );

            return next is(
                people.list_people(age_min => 23, age_max => 30),
                (vals[:3], 0, null, 3)::people.people_page,
                'can filter by age, bounds are inclusive' 
            );

            return next is(
                people.list_people(patronymic_ => ''),
                (vals[3:3], 0, null, 1)::people.people_page,
                'can filter by empty patronymic' 
            );

            return next is(
                people.list_people(
                    name_       => 'Alexandra',
                    surname_    => 'Ivanova',
                    patronymic_ => 'Alexeyevna',
                    threshold   => 0
                ),
                (array[vals[2], vals[1], vals[5], vals[4]], 0, null, 4)::people.people_page,
                'can search by similarity of name, surname, patronymic' 
            );
         
            return next is(
                people.list_people(offset_ => 1, limit_ => 2),
                (vals[2:3], 1, 2, 5)::people.people_page,
                'can use offset and limit' 
            );
        end;
    $test$
    language plpgsql;
end
$do$;
commit;
