// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package api

import (
	"github.com/google/uuid"
)

// Defines values for Sex.
const (
	Female Sex = "female"
	Male   Sex = "male"
)

// Age defines model for Age.
type Age = int

// CountryCode Country code by ISO 3166-1 alpha-2
type CountryCode = string

// PaginationOffsetLimit defines model for PaginationOffsetLimit.
type PaginationOffsetLimit struct {
	CurrentLimit  int `json:"current_limit"`
	CurrentOffset int `json:"current_offset"`
	TotalItems    int `json:"total_items"`
}

// PersonBase defines model for PersonBase.
type PersonBase struct {
	Name       *string `json:"name,omitempty"`
	Patronymic *string `json:"patronymic,omitempty"`
	Surname    *string `json:"surname,omitempty"`
}

// PersonFull defines model for PersonFull.
type PersonFull struct {
	Age  Age    `json:"age"`
	Name string `json:"name"`

	// Nationality Country code by ISO 3166-1 alpha-2
	Nationality CountryCode `json:"nationality"`
	Patronymic  string      `json:"patronymic"`
	Sex         Sex         `json:"sex"`
	Surname     string      `json:"surname"`
}

// PersonFullWithID defines model for PersonFullWithID.
type PersonFullWithID struct {
	Age  Age    `json:"age"`
	Id   UUID   `json:"id"`
	Name string `json:"name"`

	// Nationality Country code by ISO 3166-1 alpha-2
	Nationality CountryCode `json:"nationality"`
	Patronymic  string      `json:"patronymic"`
	Sex         Sex         `json:"sex"`
	Surname     string      `json:"surname"`
}

// PersonPage defines model for PersonPage.
type PersonPage struct {
	Pagination PaginationOffsetLimit `json:"pagination"`
	People     []PersonFullWithID    `json:"people"`
}

// PersonPartial defines model for PersonPartial.
type PersonPartial struct {
	Age  *Age    `json:"age,omitempty"`
	Name *string `json:"name,omitempty"`

	// Nationality Country code by ISO 3166-1 alpha-2
	Nationality *CountryCode `json:"nationality,omitempty"`
	Patronymic  *string      `json:"patronymic,omitempty"`
	Sex         *Sex         `json:"sex,omitempty"`
	Surname     *string      `json:"surname,omitempty"`
}

// PersonPostData defines model for PersonPostData.
type PersonPostData struct {
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Surname    string `json:"surname"`
}

// Sex defines model for Sex.
type Sex string

// UUID defines model for UUID.
type UUID = uuid.UUID

// PersonID defines model for personID.
type PersonID = UUID

// PersonListParams defines parameters for PersonList.
type PersonListParams struct {
	// Name Part of Person's name (case-insensitive)
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Surname Part of Person's surname (case-insensitive)
	Surname *string `form:"surname,omitempty" json:"surname,omitempty"`

	// Patronymic Part of Person's patronymic (case-insensitive, empty for no patronymic)
	Patronymic *string `form:"patronymic,omitempty" json:"patronymic,omitempty"`

	// AgeMin Minimum for Person's age
	AgeMin *Age `form:"age_min,omitempty" json:"age_min,omitempty"`

	// AgeMax Maximum for Person's age
	AgeMax *Age `form:"age_max,omitempty" json:"age_max,omitempty"`

	// Nationality Person's nationality (ISO 3166-1 alpha-2 code)
	Nationality *CountryCode `form:"nationality,omitempty" json:"nationality,omitempty"`

	// Sex Person's sex
	Sex *Sex `form:"sex,omitempty" json:"sex,omitempty"`

	// Offset The number of records to skip
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`

	// Limit The numbers of records to return (all if 0)
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`
}

// PersonPostJSONRequestBody defines body for PersonPost for application/json ContentType.
type PersonPostJSONRequestBody = PersonPostData

// PersonPatchJSONRequestBody defines body for PersonPatch for application/json ContentType.
type PersonPatchJSONRequestBody = PersonPartial

// PersonPutJSONRequestBody defines body for PersonPut for application/json ContentType.
type PersonPutJSONRequestBody = PersonFull
