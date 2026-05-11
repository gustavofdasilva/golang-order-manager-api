package config

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	API_PORT                         string
	API_VERSION                      string
	DB_PORT                          string
	DB_HOST                          string
	DB_USER                          string
	DB_PASS                          string
	DB_NAME                          string
	DB_SSLMODE                       string
	SECRET_KEY                       string
	TOKEN_EXPIRATION_MINUTES         int
	REFRESH_TOKEN_EXPIRATION_MINUTES int
	MAX_CONNECTIONS                  int
	MAX_IDLE_CONNECTIONS             int
	CONNECTION_MAX_LIFETIME_MINUTES  int
)

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	API_PORT = os.Getenv("API_PORT")
	API_VERSION = os.Getenv("API_VERSION")
	SECRET_KEY = os.Getenv("SECRET_KEY")
	DB_PORT = os.Getenv("DB_PORT")
	DB_HOST = os.Getenv("DB_HOST")
	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_NAME = os.Getenv("DB_NAME")
	DB_SSLMODE = os.Getenv("DB_SSLMODE")
	MAX_CONNECTIONS, err = strconv.Atoi(os.Getenv("MAX_CONNECTIONS"))
	if err != nil {
		slog.Error("Error parsing MAX_CONNECTIONS", "err", err)
		MAX_CONNECTIONS = 10 // default to 10
	}
	MAX_IDLE_CONNECTIONS, err = strconv.Atoi(os.Getenv("MAX_IDLE_CONNECTIONS"))
	if err != nil {
		slog.Error("Error parsing MAX_IDLE_CONNECTIONS", "err", err)
		MAX_IDLE_CONNECTIONS = 5 // default to 5
	}
	CONNECTION_MAX_LIFETIME_MINUTES, err = strconv.Atoi(os.Getenv("CONNECTION_MAX_LIFETIME_MINUTES"))
	if err != nil {
		slog.Error("Error parsing CONNECTION_MAX_LIFETIME_MINUTES", "err", err)
		CONNECTION_MAX_LIFETIME_MINUTES = 5 // default to 5 minutes
	}

	TOKEN_EXPIRATION_MINUTES, err = strconv.Atoi(os.Getenv("TOKEN_EXPIRATION_MINUTES"))
	if err != nil {
		slog.Error("Error parsing TOKEN_EXPIRATION_MINUTES", "err", err)
		TOKEN_EXPIRATION_MINUTES = 60 // default to 60 minutes
	}

	REFRESH_TOKEN_EXPIRATION_MINUTES, err = strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRATION_MINUTES"))
	if err != nil {
		slog.Error("Error parsing REFRESH_TOKEN_EXPIRATION_MINUTES", "err", err)
		REFRESH_TOKEN_EXPIRATION_MINUTES = 7 * 24 * 60 // default to 7 days in minutes
	}

	LogEnv()
}

func LogEnv() {
	slog.Info("API_PORT", "value", API_PORT)
	slog.Info("API_VERSION", "value", API_VERSION)
	slog.Info("DB_PORT", "value", DB_PORT)
	slog.Info("DB_HOST", "value", DB_HOST)
	slog.Info("DB_USER", "value", DB_USER)
	slog.Info("DB_NAME", "value", DB_NAME)
	slog.Info("DB_SSLMODE", "value", DB_SSLMODE)
	slog.Info("TOKEN_EXPIRATION_MINUTES", "value", TOKEN_EXPIRATION_MINUTES)
	slog.Info("REFRESH_TOKEN_EXPIRATION_MINUTES", "value", REFRESH_TOKEN_EXPIRATION_MINUTES)
	slog.Info("MAX_CONNECTIONS", "value", MAX_CONNECTIONS)
	slog.Info("MAX_IDLE_CONNECTIONS", "value", MAX_IDLE_CONNECTIONS)
	slog.Info("CONNECTION_MAX_LIFETIME_MINUTES", "value", CONNECTION_MAX_LIFETIME_MINUTES)
}
