package domain

import (
	"regexp"

	"github.com/google/uuid"
)

type Sex string

const (
	Male   Sex = "male"
	Female Sex = "female"
)

func (s Sex) Valid() bool {
	return (s == Male) || (s == Female)
}

type Nationality string

var NationalityPattern = regexp.MustCompile("^[A-Z]{2}$")

func (n Nationality) Valid() bool {
	return NationalityPattern.Match([]byte(n))
}

type Person struct {
	Name        string
	Surname     string
	Patronymic  string
	Nationality Nationality
	Sex         Sex
	Age         int
	ID          uuid.UUID
}

func (p Person) GetID() uuid.UUID {
	return p.ID
}

type PersonPartial struct {
	Name        *string
	Surname     *string
	Patronymic  *string
	Nationality *string
	Sex         *Sex
	Age         *int
}

type PersonFilter struct {
	Name        *string
	Surname     *string
	Patronymic  *string
	Nationality *Nationality
	Sex         *Sex
	AgeMin      *int
	AgeMax      *int
	Threshold   *float32
}

type PaginationFilter struct {
	Offset int
	Limit  int
}

type Page[T any] struct {
	Items         []T
	CurrentLimit  int
	CurrentOffset int
	TotalItems    int
}
