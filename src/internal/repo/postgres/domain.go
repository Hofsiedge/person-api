package postgres

import (
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type Person struct {
	PersonID    pgtype.UUID `db:"person_id"`
	Name        pgtype.Text `db:"name"`
	Surname     pgtype.Text `db:"surname"`
	Patronymic  pgtype.Text `db:"patronymic"`
	Age         pgtype.Int8 `db:"age"`
	Sex         pgtype.Text `db:"sex"`
	Nationality pgtype.Text `db:"nationality"`
}

// convert Person to domain.Person
func (p Person) ToAbstract() domain.Person {
	return domain.Person{
		Name:        p.Name.String,
		Surname:     p.Surname.String,
		Patronymic:  p.Patronymic.String,
		Nationality: p.Nationality.String,
		Sex:         domain.Sex(p.Sex.String),
		Age:         int(p.Age.Int64),
		ID:          p.PersonID.Bytes,
	}
}

// convert domain.Person to Person
func ToConcrete(person domain.Person) Person {
	return Person{
		PersonID:    pgtype.UUID{Bytes: person.ID, Valid: true},
		Name:        pgtype.Text{String: person.Name, Valid: true},
		Surname:     pgtype.Text{String: person.Surname, Valid: true},
		Patronymic:  pgtype.Text{String: person.Patronymic, Valid: true},
		Age:         pgtype.Int8{Int64: int64(person.Age), Valid: true},
		Sex:         pgtype.Text{String: string(person.Sex), Valid: true},
		Nationality: pgtype.Text{String: person.Nationality, Valid: true},
	}
}
