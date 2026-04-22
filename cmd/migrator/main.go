package main

import (
	"log/slog"

	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	"github.com/gladinov/effective_mobile_test_assignment/internal/migrator"

	"github.com/gladinov/effective_mobile_test_assignment/utils/logg"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.MustInitMigratorConfig()

	logger := logg.NewLogger(cfg.Env)

	err := migrator.Migrate(logger, cfg)
	if err != nil {
		logger.Error("failed to migrate", slog.Any("error", err))
	}
	logger.Info("migrations postgres applied")
}
