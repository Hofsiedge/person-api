package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type PostgresConfig struct {
	ConnString string `env:"DB_CONN" env-description:"Postgres connection string" env-required:"true"`
}

type CompleterConfig struct {
	CompleterToken string `env:"COMPLETER_TOKEN" env-description:"API token for filler services"`
	AgifyURL       string `env:"AGIFY_URL"       env-required:"true"`
	GenderizeURL   string `env:"GENDERIZE_URL"   env-required:"true"`
	NationalizeURL string `env:"NATIONALIZE_URL" env-required:"true"`
}

type ServerConfig struct {
	Debug       bool          `env:"DEBUG"         env-default:"false"`
	ReadTimout  time.Duration `env:"TIMEOUT_READ"  env-required:"true"`
	WriteTimout time.Duration `env:"TIMEOUT_WRITE" env-required:"true"`
	//nolint:tagalign
	LogLevel slog.Level `env:"LOG_LEVEL" env-required:"true" env-description:"DEBUG/INFO/WARNING/ERROR"`
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
