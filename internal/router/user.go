package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"
	"golang-order-manager-api/internal/services"

	"github.com/labstack/echo/v4"
)

func initUserRouter(e *echo.Group, userService services.UserService) {
	userPath := e.Group("/users")

	userHandler := handlers.NewUserHandler(userService)

	userPath.GET("/me", userHandler.GetUserInfo, middleware.CheckAuth)
	userPath.DELETE("/me", userHandler.DeleteUser, middleware.CheckAuth)
	userPath.PATCH("/me", userHandler.UpdateUser, middleware.CheckAuth)
}
