package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var (
			code    = http.StatusInternalServerError
			message = "internal server error"
		)

		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code

			switch msg := he.Message.(type) {
			case string:
				message = msg
			case error:
				message = msg.Error()
			default:
				message = "unexpected error"
			}
		}

		if code >= 500 {
			logger.Error(
				"http error",
				slog.Int("status", code),
				slog.Any("error", err),
			)
		}
		// TODO: почему нет обработки?
		_ = c.JSON(code, ErrorResponse{
			Error: message,
		})
	}
}
