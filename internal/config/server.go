package config

type Server struct {
	Host string `env:"HOST" env-required:"true"`
	Port string `env:"PORT" env-required:"true"`
}

func (s Server) GetServerAddress() string {
	return getAddress(s.Host, s.Port)
}
