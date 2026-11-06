package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"
	"golang-order-manager-api/internal/services"

	"github.com/labstack/echo/v4"
)

func initOrderRouter(e *echo.Group, orderService services.OrderService) {
	orders := e.Group("/orders", middleware.CheckAuth)

	orderHandler := handlers.NewOrderHandler(orderService)
	orders.GET("", orderHandler.ListOrders)
	orders.GET("/:id", orderHandler.GetOrder)
	orders.POST("", orderHandler.CreateOrder)
	orders.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
}
