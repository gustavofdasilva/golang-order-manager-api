package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func initAuthRouter(e *echo.Group) {
	authPath := e.Group("/auth")

	authPath.POST("/register", handlers.Register)
	authPath.POST("/login", handlers.Login)
	authPath.POST("/refresh", handlers.RefreshToken)
	authPath.POST("/logout", handlers.Logout, middleware.CheckAuth)
	authPath.POST("/logout-all", handlers.LogoutAll, middleware.CheckAuth)
}
