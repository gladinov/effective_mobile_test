package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/closer"
	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	"github.com/gladinov/effective_mobile_test_assignment/utils/logg"
)

type App struct {
	config      config.ServiceConfig
	logger      *slog.Logger
	diContainer *diContainer
	httpServer  *http.Server
}

func New() *App {
	a := &App{}

	a.initDeps()

	return a
}

func (a *App) initDeps() {
	inits := []func(){
		a.initConfig,
		a.initLogger,
		a.initDiContainer,
		a.initHTTPServer,
	}

	for _, fn := range inits {
		fn()
	}
}

func (a *App) initConfig() {
	a.config = config.MustInitServiceConfig()
}

func (a *App) initLogger() {
	a.logger = logg.NewLogger(a.config.Env)
}

func (a *App) initDiContainer() {
	a.diContainer = newDIContainer(a.logger, a.config)
}

func (a *App) initHTTPServer() {
	a.httpServer = &http.Server{
		Addr:         a.config.Server.GetServerAddress(),
		Handler:      a.diContainer.Handler().Routes(),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.logger.InfoContext(ctx, "start app",
		slog.String("env", a.config.Env),
		slog.String("server_host", a.config.Server.Host),
		slog.String("server_port", a.config.Server.Port))

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("server error", slog.Any("error", err))
		}
	}()

	// TODO: не застрянем ли мы здесь навсегда при ошибке в горутна выше
	<-ctx.Done()
	a.logger.Info("shutdown signal received")

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("shutdown server error", slog.Any("error", err))
	}

	a.logger.Info("server stop")

	closerCtx, closerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closerCancel()

	if err := closer.CloseAll(closerCtx); err != nil {
		slog.Error("resource close error", slog.Any("error", err))
	}

	return nil
}
