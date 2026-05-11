package middleware

import (
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/security"
	"log/slog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func CheckAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.ErrUnauthorized
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := security.ParseToken(token, config.SECRET_KEY)
		if err != nil {
			slog.Error("Failed to parse token", slog.Any("err", err))
			return echo.ErrUnauthorized
		}

		c.Set("userID", uuid.MustParse(claims.UserID))

		err = next(c)
		return err
	}
}
