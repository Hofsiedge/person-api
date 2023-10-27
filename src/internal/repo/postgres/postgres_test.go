package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/Hofsiedge/person-api/internal/repo/postgres"
	"github.com/Hofsiedge/person-api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v3"
)

type testCaseData[I any, O comparable] struct {
	input           I
	expect          O
	error           error
	setExpectations func(mock pgxmock.PgxPoolIface)
	name            string
}

type (
	funcWrapper[I any, O comparable] func(pgxmock.PgxPoolIface, I) (O, error)
	procWrapper[I any]               func(pgxmock.PgxPoolIface, I) error
)

func testFunction[I any, O comparable](t *testing.T, testCases []testCaseData[I, O], wrapper funcWrapper[I, O]) {
	t.Helper()

	for _, test := range testCases {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatal(err)
		}

		test.setExpectations(mock)
		out, err := wrapper(mock, test.input)

		switch {
		case test.error != nil || err != nil:
			if !errors.Is(err, test.error) {
				t.Errorf("%s - error mismatch: expected %v, got %v", test.name, test.error, err)
			}
		case !reflect.DeepEqual(test.expect, out):
			t.Errorf("%s - result mismatch: expected %v, got %v", test.name, test.expect, out)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

		mock.Close()
	}
}

func testProcedure[I any](t *testing.T, testCases []testCaseData[I, struct{}], wrapper procWrapper[I]) {
	t.Helper()

	for _, test := range testCases {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatal(err)
		}

		test.setExpectations(mock)

		err = wrapper(mock, test.input)

		if !errors.Is(err, test.error) {
			t.Errorf("%s - error mismatch: expected %v, got %v", test.name, test.error, err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

		mock.Close()
	}
}

//nolint:funlen
func TestCreate(t *testing.T) {
	t.Parallel()

	person := utils.MakePerson()
	validPgUUID := pgtype.UUID{Bytes: person.ID, Valid: true}
	invalidPgUUID := pgtype.UUID{Bytes: [16]byte{}, Valid: false}

	personEmptyName := person
	personEmptyName.Name = ""

	personEmptySurname := person
	personEmptySurname.Surname = ""

	personNegativeAge := person
	personNegativeAge.Age = -10

	//nolint:exhaustruct
	testCases := []testCaseData[domain.Person, uuid.UUID]{
		{
			name: "valid args",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select people.create_person`).
					WithArgs(
						person.Name, person.Surname, person.Patronymic,
						person.Age, person.Sex, person.Nationality).
					WillReturnRows(
						mock.NewRows([]string{"person_id"}).
							AddRow(validPgUUID))
			},
			input:  person,
			expect: person.ID,
		},
		{
			name: "invalid name - should not be empty string",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select people.create_person`).
					WithArgs("", person.Surname, person.Patronymic,
						person.Age, person.Sex, person.Nationality).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			input: personEmptyName,
			error: repo.ErrArgument,
		},
		{
			name: "invalid surname - should not be empty string",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select people.create_person`).
					WithArgs(person.Name, "", person.Patronymic,
						person.Age, person.Sex, person.Nationality).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			input: personEmptySurname,
			error: repo.ErrArgument,
		},
		{
			name: "invalid age - should not be negative",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select people.create_person`).
					WithArgs(person.Name, person.Surname, person.Patronymic,
						personNegativeAge.Age, person.Sex, person.Nationality).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			input: personNegativeAge,
			error: repo.ErrArgument,
		},
		{
			name: "unexpected error",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select people.create_person`).
					WithArgs(
						person.Name, person.Surname, person.Patronymic,
						person.Age, person.Sex, person.Nationality).
					WillReturnRows(mock.NewRows([]string{"person_id"}).AddRow(invalidPgUUID))
			},
			input: person,
			error: repo.ErrUnexpected,
		},
	}

	wrapper := func(mock pgxmock.PgxPoolIface, person domain.Person) (uuid.UUID, error) {
		repo := postgres.PeopleFromPgxPoolInterface(mock)

		return repo.Create(context.Background(), person) //nolint:wrapcheck
	}
	testFunction[domain.Person, uuid.UUID](t, testCases, wrapper)
}

func TestGet(t *testing.T) {
	t.Parallel()

	person := utils.MakePerson()
	pgPerson := postgres.ToConcrete(person)

	//nolint:exhaustruct
	testCases := []testCaseData[uuid.UUID, domain.Person]{
		{
			name: "valid",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select \* from people.get_person`).
					WithArgs(person.ID).
					WillReturnRows(
						mock.NewRows([]string{
							"person_id", "name", "surname", "patronymic",
							"age", "sex", "nationality",
						}).AddRow(
							pgPerson.PersonID, pgPerson.Name, pgPerson.Surname,
							pgPerson.Patronymic, pgPerson.Age, pgPerson.Sex,
							pgPerson.Nationality,
						),
					)
			},
			input:  person.ID,
			expect: person,
		},
		{
			name: "not found",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select \* from people.get_person`).
					WithArgs(person.ID).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NoDataFound})
			},
			input: person.ID,
			error: repo.ErrNotFound,
		},
		{
			name: "unexpected error",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`^select \* from people.get_person`).
					WithArgs(person.ID).
					WillReturnRows(mock.NewRows([]string{"unexpected_column"}).AddRow(person.ID))
			},
			input: person.ID,
			error: repo.ErrUnexpected,
		},
	}

	wrapper := func(mock pgxmock.PgxPoolIface, personID uuid.UUID) (domain.Person, error) {
		repo := postgres.PeopleFromPgxPoolInterface(mock)

		return repo.GetByID(context.Background(), personID) //nolint:wrapcheck
	}
	testFunction[uuid.UUID, domain.Person](t, testCases, wrapper)
}

//nolint:exhaustruct
func TestDelete(t *testing.T) {
	t.Parallel()

	personID := uuid.New()

	testCases := []testCaseData[uuid.UUID, struct{}]{
		{
			name: "delete non-existent person",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`^select people.delete_person`).
					WithArgs(personID).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NoDataFound})
			},
			input: personID,
			error: repo.ErrNotFound,
		},
		{
			name: "delete existing person",
			setExpectations: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`^select people.delete_person`).
					WithArgs(personID).
					WillReturnResult(pgxmock.NewResult("SELECT", 1))
			},
			input: personID,
		},
	}

	wrapper := func(mock pgxmock.PgxPoolIface, id uuid.UUID) error {
		return postgres.PeopleFromPgxPoolInterface(mock).Delete(context.Background(), id) //nolint:wrapcheck
	}
	testProcedure[uuid.UUID](t, testCases, wrapper)
}
