package logg

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func NewLogger(env string) *slog.Logger {
	var (
		level   slog.Level
		handler slog.Handler
	)

	switch env {
	case envLocal:
		level = slog.LevelDebug
		base := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		handler = New(base)

	case envDev:
		level = slog.LevelDebug
		base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		handler = New(base)

	case envProd:
		level = slog.LevelInfo
		base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		handler = New(base)

	default:
		level = slog.LevelInfo
		base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		handler = New(base)

	}
	return slog.New(handler)
}
