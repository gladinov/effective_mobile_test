package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServiceConfig struct {
	Env      string `env:"ENV" env-required:"true"`
	Server   Server
	Postgres Postgres
	Timeouts Timeouts
}

type Timeouts struct {
	DbQueryTimeout   time.Duration `env:"DB_QUERY_TIMEOUT" env-required:"true"`
	RequestTimeout   time.Duration `env:"REQUEST_TIMEOUT" env-required:"true"`
	AppCloseTimeout  time.Duration `env:"APP_CLOSE_TIMEOUT" env-required:"true"`
	DbConnectTimeout time.Duration `env:"DB_CONNECT_TIMEOUT" env-required:"true"`
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
