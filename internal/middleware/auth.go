package middleware

import (
	"golang-order-manager-api/internal/config"
	auth "golang-order-manager-api/internal/services"
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

		authService := auth.NewAuthService(nil, config.SECRET_KEY)

		claims, err := authService.ParseToken(token)
		if err != nil {
			slog.Error("Failed to parse token", slog.Any("err", err))
			return echo.ErrUnauthorized
		}

		c.Set("userID", uuid.MustParse(claims.UserID))

		err = next(c)
		return err
	}
}
