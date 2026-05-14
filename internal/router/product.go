package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func initProductRouter(e *echo.Group) {
	products := e.Group("/products")

	products.GET("", handlers.ListProducts)
	products.GET("/:id", handlers.GetProduct)
	products.POST("", handlers.CreateProduct, middleware.CheckAuth)
	products.PATCH("/:id", handlers.UpdateProduct, middleware.CheckAuth)
	products.DELETE("/:id", handlers.DeleteProduct, middleware.CheckAuth)
}
