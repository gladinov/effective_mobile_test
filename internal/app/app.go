package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	_ "github.com/gladinov/effective_mobile_test_assignment/docs"
	"github.com/gladinov/effective_mobile_test_assignment/internal/closer"
	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	httperrors "github.com/gladinov/effective_mobile_test_assignment/internal/http/errors"
	mw "github.com/gladinov/effective_mobile_test_assignment/internal/http/middleware"
	"github.com/gladinov/effective_mobile_test_assignment/utils/logg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type App struct {
	config      config.ServiceConfig
	logger      *slog.Logger
	diContainer *diContainer
	router      http.Handler
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
		a.initRouter,
		a.initHTTPServer,
	}

	for _, fn := range inits {
		fn()
	}
}

func (a *App) initRouter() {
	router := echo.New()

	router.Use(middleware.CORS())
	router.Use(mw.LoggerMiddleWare(a.logger))
	router.HTTPErrorHandler = httperrors.HTTPErrorHandler(a.logger)
	// TODO: почему не в RegisterRoutes?
	router.GET("/swagger/*", echoSwagger.WrapHandler)
	a.diContainer.Handler().RegisterRoutes(router)

	a.router = router
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
		Addr:              a.config.Server.GetServerAddress(),
		Handler:           a.router,
		ReadHeaderTimeout: a.config.Server.ReadHeaderTimeout,
		WriteTimeout:      a.config.Server.WriteTimeout,
		ReadTimeout:       a.config.Server.ReadTimeout,
		IdleTimeout:       a.config.Server.IdleTimeout,
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.logger.InfoContext(ctx, "start app",
		slog.String("env", a.config.Env),
		slog.String("server_host", a.config.Server.Host),
		slog.String("server_port", a.config.Server.Port))

	errCh := make(chan error, 1)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("server error", slog.Any("error", err))
			errCh <- err
		}
	}()

	// TODO: не застрянем ли мы здесь навсегда при ошибке в горутна выше
	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case err := <-errCh:
		a.logger.ErrorContext(ctx, "server stopped with error", slog.Any("error", err))
	}

	// Паттерн "двойной Ctrl+C":
	// первый сигнал запускает graceful shutdown,
	// второй мгновенно завершает процесс.
	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.config.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("shutdown server error", slog.Any("error", err))
	}

	a.logger.Info("server stop")

	closerCtx, closerCancel := context.WithTimeout(context.Background(), a.config.Timeouts.AppCloseTimeout)
	defer closerCancel()

	if err := closer.CloseAll(closerCtx); err != nil {
		a.logger.Error("resource close error", slog.Any("error", err))
	}

	return nil
}
