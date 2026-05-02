package database

import (
	"database/sql"
	"fmt"
	"golang-order-manager-api/internal/config"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	dbHost := config.DB_HOST
	dbPort := config.DB_PORT
	dbUser := config.DB_USER
	dbPass := config.DB_PASS
	dbName := config.DB_NAME
	dbSslMode := config.DB_SSLMODE
	maxOpenConn := config.MAX_CONNECTIONS
	maxIdleConn := config.MAX_IDLE_CONNECTIONS
	connMaxLifetime := config.CONNECTION_MAX_LIFETIME_MINUTES

	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", dbHost, dbUser, dbPass, dbName, dbPort, dbSslMode))

	db.SetMaxOpenConns(maxOpenConn)
	db.SetMaxIdleConns(maxIdleConn)
	db.SetConnMaxLifetime(time.Minute * time.Duration(connMaxLifetime))

	if err != nil {
		panic(err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Successfully connected to database")
}

func GetDB() *sql.DB {
	return db
}
