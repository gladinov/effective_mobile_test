package config

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyHost     error = errors.New("empty host in config")
	ErrEmptyUser     error = errors.New("empty user in config")
	ErrEmptyPassword error = errors.New("empty password in config")
	ErrEmptyDBName   error = errors.New("empty dbname in config")
	ErrEmptyPort     error = errors.New("empty port in config")
)

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Dbname   string `env:"POSTGRES_DB" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	PgUser   string `env:"PGUSER" env-required:"true"`
	SslMode  string `env:"SSLMODE" env-required:"true"`
}

func (p *Postgres) GetAdress() string {
	return getAddress(p.Host, p.Port)
}

func (p *Postgres) GetDSN() (string, error) {
	if p.Host == "" {
		return "", ErrEmptyHost
	}
	if p.User == "" {
		return "", ErrEmptyUser
	}
	if p.Password == "" {
		return "", ErrEmptyPassword
	}
	if p.Dbname == "" {
		return "", ErrEmptyDBName
	}
	if p.Port == "" {
		return "", ErrEmptyPort
	}
	host := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=%s", p.User, p.Password, p.Host, p.Port, p.Dbname, p.SslMode)
	return host, nil
}
