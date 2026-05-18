package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func initUserRouter(e *echo.Group) {
	userPath := e.Group("/users")

	userPath.GET("/me", handlers.GetUserInfo, middleware.CheckAuth)
	userPath.DELETE("/me", handlers.DeleteUser, middleware.CheckAuth)
	userPath.PATCH("/me", handlers.UpdateUser, middleware.CheckAuth)
}
