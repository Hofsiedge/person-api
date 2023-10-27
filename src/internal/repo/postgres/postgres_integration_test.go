package postgres_test

import (
	"context"
	"flag"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Hofsiedge/person-api/internal/config"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/Hofsiedge/person-api/internal/repo/postgres"
	"github.com/Hofsiedge/person-api/internal/utils"
)

func TestIntegration(t *testing.T) {
	if !runIntegrationTests {
		t.SkipNow()
	}

	t.Parallel()

	person := utils.MakePerson()

	var err error

	person.ID, err = people.Create(context.Background(), person)
	if err != nil {
		t.Fatalf("could not create a Person: %v", err)
	}

	result, err := people.GetByID(context.Background(), person.ID)
	if err != nil {
		t.Errorf("could not get a Person: %v", err)
	}

	if !reflect.DeepEqual(person, result) {
		t.Errorf("mismatch in GetByID result: expected %v, got %v", person, result)
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
		cfg, err := config.Read[config.DBConfig]()
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
