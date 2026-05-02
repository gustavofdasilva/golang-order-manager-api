package main

import (
	"fmt"
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/middleware"
	"golang-order-manager-api/internal/router"
	"golang-order-manager-api/pkg/database"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	config.InitConfig()
	database.InitDB()
	e := echo.New()

	api := e.Group(fmt.Sprintf("/api/v%s", config.API_VERSION))

	api.Use(middleware.LogRequest)
	api.Use(middleware.CORSConfig())

	router.InitRouter(api)

	port := strings.TrimLeft(config.API_PORT, ":")
	e.Logger.Fatal(e.Start(fmt.Sprintf("127.0.0.1:%s", port)))
}
