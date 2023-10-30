package postgres_test

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Hofsiedge/person-api/internal/config"
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/Hofsiedge/person-api/internal/repo/postgres"
	"github.com/Hofsiedge/person-api/internal/utils"
)

//nolint:funlen,cyclop
func TestIntegration(t *testing.T) {
	if !runIntegrationTests {
		t.SkipNow()
	}

	t.Parallel()

	person := utils.MakePerson()

	var err error

	// create (POST)
	person.ID, err = people.Create(context.Background(), person)
	if err != nil {
		t.Fatalf("could not create a Person: %v", err)
	}

	// get
	result, err := people.GetByID(context.Background(), person.ID)
	if err != nil {
		t.Errorf("could not get a Person: %v", err)
	} else if !reflect.DeepEqual(person, result) {
		t.Errorf("mismatch in GetByID result: expected %v, got %v", person, result)
	}

	// replace (PUT)
	anotherPerson := utils.MakePerson()
	if err = people.FullUpdate(context.Background(), person.ID, anotherPerson); err != nil {
		t.Errorf("could not replace a Person: %v", err)
	} else {
		anotherPerson.ID = person.ID

		result, err = people.GetByID(context.Background(), person.ID)
		if err != nil {
			t.Fatalf("could not get a Person after replacement: %v", err)
		}

		if !reflect.DeepEqual(result, anotherPerson) {
			t.Errorf("replacement was not saved: expected %v, got %v", anotherPerson, result)
		}
	}

	// update
	yetAnotherPerson := utils.MakePerson()
	partial := domain.PersonPartial{Name: &yetAnotherPerson.Name} //nolint:exhaustruct

	if err = people.PartialUpdate(context.Background(), person.ID, partial); err != nil {
		t.Errorf("could not update a Person: %v", err)
	} else {
		anotherPerson.Name = yetAnotherPerson.Name
		result, err = people.GetByID(context.Background(), person.ID)
		if err != nil {
			t.Fatalf("could not get a Person after update: %v", err)
		}
		if !reflect.DeepEqual(result, anotherPerson) {
			t.Errorf("update was not saved: expected %v, got %v", anotherPerson, result)
		}
	}

	// delete
	if err = people.Delete(context.Background(), person.ID); err != nil {
		t.Errorf("could not delete a Person: %v", err)
	} else {
		_, err = people.GetByID(context.Background(), person.ID)
		if !errors.Is(err, repo.ErrNotFound) {
			t.Errorf("the Person was not deleted. error: %v", err)
		}
	}

	// list
	for i := 0; i < 10; i++ {
		person = utils.MakePerson()
		if _, err = people.Create(context.Background(), person); err != nil {
			t.Fatalf("could not create a Person: %v", err)
		}
	}

	page, err := people.List(
		context.Background(),
		domain.PersonFilter{}, //nolint:exhaustruct
		domain.PaginationFilter{Offset: 0, Limit: 0},
	)
	if err != nil {
		t.Errorf("could not list Person records: %v", err)
	} else if page.CurrentLimit != 0 || page.CurrentOffset != 0 || page.TotalItems != 10 {
		t.Errorf("page mismatch: got %v", page)
	}
}

//nolint:gochecknoglobals
var (
	runIntegrationTests bool
	people              repo.PersonRepo
)

func TestMain(m *testing.M) {
	flag.BoolVar(&runIntegrationTests, "integration", false, "run integration tests or not")
	flag.Parse()

	if runIntegrationTests {
		cfg, err := config.Read[config.PostgresConfig]()
		if err != nil {
			log.Fatal(err)
		}

		people, err = postgres.New(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	os.Exit(m.Run())
}
