package main

import (
	"log/slog"
	"os"

	"github.com/gladinov/effective_mobile_test_assignment/internal/app"
)

func main() {
	a := app.New()

	if err := a.Run(); err != nil {
		slog.Error("app error", slog.Any("error", err))
		os.Exit(1)
	}
}
