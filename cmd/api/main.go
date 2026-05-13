package main

import (
	"fmt"
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/middleware"
	"golang-order-manager-api/internal/router"
	"golang-order-manager-api/pkg/database"
	"golang-order-manager-api/pkg/logger"
	"strings"

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
	e := echo.New()

	api := e.Group(fmt.Sprintf("/api/v%s", config.API_VERSION))

	api.Use(middleware.RateLimiter())
	api.Use(middleware.LogRequest)
	api.Use(middleware.CORSConfig())

	router.InitRouter(api)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	port := strings.TrimLeft(config.API_PORT, ":")
	e.Logger.Fatal(e.Start(fmt.Sprintf("127.0.0.1:%s", port)))
}
