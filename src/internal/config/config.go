package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type DBConfig struct {
	ConnString string `env:"DB_CONN" env-description:"Postgres connection string" env-required:"true"`
}

// read config from environment variables
//
// returns invalid config on error
func Read[T any]() (T, error) {
	var (
		cfg T
		err error
	)

	if err = cleanenv.ReadEnv(&cfg); err != nil {
		helpHeader := "Expected config:"
		help, descErr := cleanenv.GetDescription(&cfg, &helpHeader)

		if descErr == nil {
			err = fmt.Errorf("could not read config: %w.\n%s", err, help)
		} else {
			err = fmt.Errorf("could not read config: %w", err)
		}
	}

	return cfg, err
}
