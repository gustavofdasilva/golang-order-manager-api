package router

import (
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"
	"golang-order-manager-api/internal/services"

	"github.com/labstack/echo/v4"
)

func initProductRouter(e *echo.Group, productService services.ProductService) {
	products := e.Group("/products")

	productHandler := handlers.NewProductHandler(productService)
	products.GET("", productHandler.ListProducts)
	products.GET("/:id", productHandler.GetProduct)
	products.POST("", productHandler.CreateProduct, middleware.CheckAuth)
	products.PATCH("/:id", productHandler.UpdateProduct, middleware.CheckAuth)
	products.DELETE("/:id", productHandler.DeleteProduct, middleware.CheckAuth)
}
