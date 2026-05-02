package router

import (
	"golang-order-manager-api/internal/handlers"

	"github.com/labstack/echo/v4"
)

func initUserRouter(e *echo.Group) {
	userPath := e.Group("/auth")

	userPath.POST("/register", handlers.CreateUser)
	userPath.POST("/login", handlers.Login)
	userPath.GET("/info", handlers.GetUserInfo)
}
