package middleware

import (
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/responses"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func RateLimiter() echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(config.RATE_LIMIT_RPS),
				Burst:     config.RATE_LIMIT_BURST,
				ExpiresIn: time.Duration(config.RATE_LIMIT_EXPIRES_MINUTES) * time.Minute,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusInternalServerError, responses.ErrorResponse{Error: err.Error()})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, responses.ErrorResponse{Error: "too many requests"})
		},
	})
}
