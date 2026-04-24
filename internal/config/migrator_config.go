package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type MigratorConfig struct {
	Env            string `env:"ENV" env-required:"true"`
	MigrationsPath string `env:"MIGRATIONS_PATH" env-required:"true"`
	Postgres       Postgres
}

func MustInitMigratorConfig() MigratorConfig {
	var config MigratorConfig
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return config
}
