package migrator

import (
	"errors"
	"log/slog"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	"github.com/golang-migrate/migrate/v4"
)

func Migrate(logger *slog.Logger, cfg config.MigratorConfig) error {
	migrationsURL := "file://" + cfg.MigrationsPath + "/"

	logger.Info("migrationURL", slog.String("migrationURL", cfg.MigrationsPath))

	databaseURL, err := cfg.Postgres.GetDSN()
	if err != nil {
		return e.WrapIfErr("get dsn from config", err)
	}

	m, err := migrate.New(migrationsURL, databaseURL)
	if err != nil {
		return e.WrapIfErr("new migrate instance", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no migrations to apply")
			return nil
		}
		return e.WrapIfErr("up migrate", err)
	}
	return nil
}
