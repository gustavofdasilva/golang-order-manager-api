package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"
	"golang-order-manager-api/internal/services"

	"github.com/labstack/echo/v4"
)

func initAuthRouter(e *echo.Group, authService services.AuthService) {
	authPath := e.Group("/auth")

	authHandler := handlers.NewAuthHandler(authService)
	authPath.POST("/register", authHandler.Register)
	authPath.POST("/login", authHandler.Login)
	authPath.POST("/refresh", authHandler.RefreshToken)
	authPath.POST("/logout", authHandler.Logout, middleware.CheckAuth)
	authPath.POST("/logout-all", authHandler.LogoutAll, middleware.CheckAuth)
}
