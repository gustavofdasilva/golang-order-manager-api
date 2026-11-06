package router

import (
	"golang-order-manager-api/internal/services"

	"github.com/labstack/echo/v4"
)

type RouterServices struct {
	UserService    services.UserService
	AuthService    services.AuthService
	ProductService services.ProductService
	OrderService   services.OrderService
}

func InitRouter(e *echo.Group, routerServices RouterServices) {
	initUserRouter(e, routerServices.UserService)
	initAuthRouter(e, routerServices.AuthService)
	initProductRouter(e, routerServices.ProductService)
	initOrderRouter(e, routerServices.OrderService)
}
