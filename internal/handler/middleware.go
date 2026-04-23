package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *handler) LoggerMiddleWare(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		logg := h.logger.With(
			slog.String("component", "middleware/logger"),
		)

		req := c.Request()
		resp := c.Response()
		entry := logg.With(
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("remote_addr", req.RemoteAddr),
			slog.String("user_agent", req.UserAgent()),
		)
		start := time.Now()

		defer func() {
			var status int

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = http.StatusInternalServerError
				}
			} else {
				status = resp.Status
			}

			attrs := []any{
				slog.Int("status", status),
				slog.Int64("bytes", resp.Size),
				slog.Duration("duration", time.Since(start)),
			}

			if err != nil {
				entry.Error("request failed",
					append(attrs, slog.Any("error", err))...,
				)
			} else {
				entry.Info("request completed", attrs...)
			}
		}()
		err = next(c)

		return err
	}
}
