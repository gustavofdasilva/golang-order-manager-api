package main

import (
	"fmt"
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/middleware"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/router"
	"golang-order-manager-api/internal/services"
	"golang-order-manager-api/pkg/cache"
	"golang-order-manager-api/pkg/database"
	"golang-order-manager-api/pkg/logger"
	"strings"
	"time"

	_ "golang-order-manager-api/migrations"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Golang Order Manager API
// @version 1.0
// @description Production-inspired REST API built with Go.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	logger.Init()
	config.InitConfig()
	database.InitDB()
	cache.InitRedis()
	e := echo.New()

	api := e.Group(fmt.Sprintf("/api/v%s", config.API_VERSION))

	api.Use(middleware.RateLimiter())
	api.Use(middleware.LogRequest)
	api.Use(middleware.CORSConfig())

	db := database.GetDB()

	// repos
	userRepo := repository.NewUserRepo(db)
	authRepo := repository.NewAuthRepo(db)
	productRepo := repository.NewProductRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	orderTxFactory := repository.NewTxFactory(db)
	orderCache := cache.NewRedisOrderCache()

	// services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, authRepo, config.SECRET_KEY, time.Duration(config.REFRESH_TOKEN_EXPIRATION_MINUTES)*time.Minute)
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderTxFactory, orderRepo, orderCache)

	router.InitRouter(api, router.RouterServices{
		UserService:    userService,
		AuthService:    authService,
		ProductService: productService,
		OrderService:   orderService,
	})

	e.GET("/health", handlers.Health)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	port := strings.TrimLeft(config.API_PORT, ":")
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%s", port)))
}
