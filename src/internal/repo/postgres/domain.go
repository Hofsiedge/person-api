package postgres

import (
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/google/uuid"
)

type Person struct {
	PersonID    uuid.UUID `db:"person_id"`
	Name        string    `db:"name"`
	Surname     string    `db:"surname"`
	Patronymic  string    `db:"patronymic"`
	Age         int       `db:"age"`
	Sex         string    `db:"sex"`
	Nationality string    `db:"nationality"`
}

// convert Person to domain.Person
func (p Person) ToAbstract() domain.Person {
	return domain.Person{
		Name:        p.Name,
		Surname:     p.Surname,
		Patronymic:  p.Patronymic,
		Nationality: domain.Nationality(p.Nationality),
		Sex:         domain.Sex(p.Sex),
		Age:         p.Age,
		ID:          p.PersonID,
	}
}

// convert domain.Person to Person
func ToConcrete(person domain.Person) Person {
	return Person{
		PersonID:    person.ID,
		Name:        person.Name,
		Surname:     person.Surname,
		Patronymic:  person.Patronymic,
		Age:         person.Age,
		Sex:         string(person.Sex),
		Nationality: string(person.Nationality),
	}
}

type PersonPage struct {
	People        []Person `db:"people"`
	CurrentOffset int      `db:"current_offset"`
	CurrentLimit  int      `db:"current_limit"`
	Total         int      `db:"total"`
}

// convert PersonPage to domain.Page[domain.Person]
func (p PersonPage) ToAbstract() domain.Page[domain.Person] {
	page := domain.Page[domain.Person]{
		Items:         make([]domain.Person, len(p.People)),
		CurrentOffset: p.CurrentOffset,
		CurrentLimit:  p.CurrentLimit,
		TotalItems:    p.Total,
	}
	for i, person := range p.People {
		page.Items[i] = person.ToAbstract()
	}

	return page
}
