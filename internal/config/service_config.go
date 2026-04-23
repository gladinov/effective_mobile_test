package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServiceConfig struct {
	Env      string `env:"ENV" env-required:"true"`
	Server   Server
	Postgres Postgres
}

func MustInitServiceConfig() ServiceConfig {
	var config ServiceConfig
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return config
}

func getAddress(host string, port string) string {
	return host + ":" + port
}


