package router

import (
	"golang-order-manager-api/internal/handlers"

	"github.com/labstack/echo/v4"
)

func initAuthRouter(e *echo.Group) {
	authPath := e.Group("/auth")

	authPath.POST("/register", handlers.Register)
	authPath.POST("/login", handlers.Login)
}
