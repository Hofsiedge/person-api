package utils

import (
	"math/rand"
	"strings"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//nolint:gosec
func GenerateRandomString(minLength, maxLength uint) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"

	var builder strings.Builder

	length := minLength + (uint(rand.Int()) % (maxLength + 1 - minLength))
	for i := 0; i < int(length); i++ {
		letter := letters[rand.Int()%len(letters)]
		builder.WriteByte(letter)
	}

	return builder.String()
}

//nolint:gosec,gomnd
func MakePerson() domain.Person {
	capitalizer := cases.Title(language.Und)

	sex := domain.Male
	if rand.Float32() < 0.5 {
		sex = domain.Female
	}

	return domain.Person{
		Name:        capitalizer.String(GenerateRandomString(2, 10)),
		Surname:     capitalizer.String(GenerateRandomString(2, 20)),
		Patronymic:  capitalizer.String(GenerateRandomString(0, 10)),
		Nationality: strings.ToTitle(GenerateRandomString(2, 2)),
		Sex:         sex,
		Age:         rand.Int() % 120,
		ID:          uuid.New(),
	}
}
