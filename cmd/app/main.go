package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	"github.com/gladinov/effective_mobile_test_assignment/internal/handler"
	"github.com/gladinov/effective_mobile_test_assignment/internal/repository/postgres"
	"github.com/gladinov/effective_mobile_test_assignment/internal/service"
	"github.com/gladinov/effective_mobile_test_assignment/utils/logg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	conf := config.MustInitServiceConfig()

	logger := logg.NewLogger(conf.Env)

	logger.Info("start app",
		slog.String("env", conf.Env),
		slog.String("server_host", conf.Server.Host),
		slog.String("server_port", conf.Server.Port))

	logger.InfoContext(ctx, "create new pool")
	pool, err := postgres.NewPool(ctx, conf)
	if err != nil {
		return
	}
	logger.InfoContext(ctx, "create new storage")
	storage := postgres.NewStorage(pool)

	defer storage.Close()

	service := service.NewService(logger, storage)

	h := handler.NewHandler(logger, service)

	router := echo.New()

	router.Use(middleware.CORS())
	router.Use(h.LoggerMiddleWare)
	router.HTTPErrorHandler = handler.HTTPErrorHandler(logger)

	router.POST("/subscriptions/create", h.Create)
	router.GET("/subscriptions/get/:id", h.Get)
	router.PUT("/subscriptions/update/:id", h.Update)
	router.DELETE("/subscriptions/delete/:id", h.Delete)
	router.DELETE("/subscriptions/list", h.List)
	router.GET("/subscriptions/total", h.Total)
	// srv := http.Server{}

	// go func(){srv := http.ListenAndServe}

	// gracefulShutdown()
}
