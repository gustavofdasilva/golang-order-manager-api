package handlers

import (
	"net/http"
	"time"

	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/pkg/database"

	"github.com/labstack/echo/v4"
)

var startTime = time.Now()

type dbHealthStatus struct {
	Status          string `json:"status"`
	Error           string `json:"error,omitempty"`
	OpenConnections int    `json:"open_connections,omitempty"`
	InUse           int    `json:"in_use,omitempty"`
	Idle            int    `json:"idle,omitempty"`
	MaxOpen         int    `json:"max_open,omitempty"`
}

type healthResponse struct {
	Status   string         `json:"status"`
	Version  string         `json:"version"`
	Uptime   string         `json:"uptime"`
	Database dbHealthStatus `json:"database"`
}

// Health godoc
//
// @Summary Health check
// @Description Returns the health status of the API and its dependencies. HTTP 503 when degraded.
// @Tags Health
// @Produce json
// @Success 200 {object} healthResponse "API is healthy"
// @Failure 503 {object} healthResponse "API is degraded (database unreachable)"
// @Router /health [get]
func Health(c echo.Context) error {
	db := database.GetDB()

	dbStatus := dbHealthStatus{}
	if err := db.Ping(); err != nil {
		dbStatus.Status = "down"
		dbStatus.Error = err.Error()
	} else {
		stats := db.Stats()
		dbStatus.Status = "up"
		dbStatus.OpenConnections = stats.OpenConnections
		dbStatus.InUse = stats.InUse
		dbStatus.Idle = stats.Idle
		dbStatus.MaxOpen = stats.MaxOpenConnections
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if dbStatus.Status == "down" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, healthResponse{
		Status:   status,
		Version:  config.API_VERSION,
		Uptime:   time.Since(startTime).Truncate(time.Second).String(),
		Database: dbStatus,
	})
}
