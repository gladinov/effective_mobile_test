package config

import "time"

type Server struct {
	Host              string        `env:"HOST" env-required:"true"`
	Port              string        `env:"PORT" env-required:"true"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" env-required:"true"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" env-required:"true"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" env-required:"true"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" env-required:"true"`
	ShutdownTimeout   time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-required:"true"`
}

func (s Server) GetServerAddress() string {
	return getAddress(s.Host, s.Port)
}
