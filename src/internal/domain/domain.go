package domain

import (
	"github.com/google/uuid"
)

type Sex string

const (
	Male   Sex = "male"
	Female Sex = "female"
)

type Person struct {
	Name        string
	Surname     string
	Patronymic  string
	Nationality string
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
	Nationality *string // TODO: separate type
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
