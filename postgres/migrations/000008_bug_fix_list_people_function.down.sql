begin;

drop function people.list_people(text, text, text, int, int, people.sex, char(2), real, int, int);

-- the version from migration #7
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

commit;
