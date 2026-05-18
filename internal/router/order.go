package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func initOrderRouter(e *echo.Group) {
	orders := e.Group("/orders", middleware.CheckAuth)

	orders.GET("", handlers.ListOrders)
	orders.GET("/:id", handlers.GetOrder)
	orders.POST("", handlers.CreateOrder)
	orders.PATCH("/:id/status", handlers.UpdateOrderStatus)
}
