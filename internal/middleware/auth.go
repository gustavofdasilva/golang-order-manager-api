package middleware

import (
	"golang-order-manager-api/internal/services"

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

		valid, err := services.VerifyToken(token)
		if err != nil || !valid {
			return echo.ErrUnauthorized
		}

		err = next(c)
		return err
	}
}
