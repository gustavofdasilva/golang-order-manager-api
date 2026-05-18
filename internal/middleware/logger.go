package middleware

import (
	"golang-order-manager-api/pkg/logger"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func LogRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		requestID := c.Request().Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Response().Header().Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		err := next(c)

		status := c.Response().Status
		latencyMs := time.Since(start).Milliseconds()

		attrs := []slog.Attr{
			slog.String(logger.KeyRequestID, requestID),
			slog.String(logger.KeyMethod, c.Request().Method),
			slog.String(logger.KeyPath, c.Request().URL.Path),
			slog.Int(logger.KeyStatus, status),
			slog.Int64(logger.KeyLatencyMs, latencyMs),
			slog.String(logger.KeyIP, c.RealIP()),
			slog.String(logger.KeyUserAgent, c.Request().UserAgent()),
			slog.Int64(logger.KeyBytesIn, c.Request().ContentLength),
			slog.Int64(logger.KeyBytesOut, c.Response().Size),
		}

		q := c.Request().URL.RawQuery
		if q != "" {
			attrs = append(attrs, slog.String(logger.KeyQuery, q))
		}

		if err != nil {
			attrs = append(attrs, slog.String(logger.KeyError, err.Error()))
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(c.Request().Context(), level, "request", attrs...)

		return err
	}
}

func CORSConfig() echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	})
}
