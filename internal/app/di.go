package app

import (
	"context"
	"log/slog"

	"github.com/gladinov/effective_mobile_test_assignment/internal/closer"
	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	"github.com/gladinov/effective_mobile_test_assignment/internal/http/handler"
	"github.com/gladinov/effective_mobile_test_assignment/internal/repository/postgres"
	"github.com/gladinov/effective_mobile_test_assignment/internal/service"
)

type diContainer struct {
	logger  *slog.Logger
	cfg     config.ServiceConfig
	db      service.Storage
	service handler.Service
	handler handler.Handler
}

func newDIContainer(logger *slog.Logger, config config.ServiceConfig) *diContainer {
	return &diContainer{
		logger: logger,
		cfg:    config,
	}
}

func (d *diContainer) DB() service.Storage {
	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.Timeouts.DbQueryTimeout)
	defer cancel()
	if d.db == nil {
		d.logger.Info("create new pool")
		pool, err := postgres.NewPool(ctx, d.cfg)
		if err != nil {
			d.logger.Error("create postgres pool", slog.Any("error", err))
			panic(err)
		}

		d.logger.Info("create new storage")
		storage := postgres.NewStorage(pool, d.cfg.Timeouts.DbConnectTimeout)

		closer.Add("postgres DB", func(_ context.Context) error {
			return storage.Close()
		})

		d.db = storage
	}

	return d.db
}

func (d *diContainer) Service() handler.Service {
	if d.service == nil {
		d.logger.Info("create new service")
		service := service.NewService(d.DB())
		d.service = service
	}
	return d.service
}

func (d *diContainer) Handler() handler.Handler {
	if d.handler == nil {
		d.logger.Info("create new handler")
		h := handler.NewHandler(d.logger, d.Service(), d.cfg.Timeouts.RequestTimeout)
		d.handler = h
	}
	return d.handler
}
