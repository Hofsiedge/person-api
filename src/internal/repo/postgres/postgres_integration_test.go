package postgres_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Hofsiedge/person-api/internal/config"
	"github.com/Hofsiedge/person-api/internal/domain"
	"github.com/Hofsiedge/person-api/internal/repo"
	"github.com/Hofsiedge/person-api/internal/repo/postgres"
)

func TestIntegration(t *testing.T) {
	if !runIntegrationTests {
		t.SkipNow()
	}

	t.Parallel()

	person := domain.Person{
		Name:        "Name",
		Surname:     "Surname",
		Patronymic:  "Patronymic",
		Nationality: "NA",
		Sex:         domain.Male,
		Age:         100,
		ID:          [16]byte{},
	}

	var err error

	person.ID, err = people.Create(context.Background(), person)
	if err != nil {
		t.Fatalf("could not create a Person: %v", err)
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
