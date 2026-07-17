package database

import (
	"database/sql"
	"fmt"
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/migrations"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

	log.Println("Successfully connected to database")

	RunMigrations(db)
}

func GetDB() *sql.DB {
	return db
}

func RunMigrations(db *sql.DB) {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("error to create iofs driver: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("error to create db driver: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		log.Fatalf("error to start migrations: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("error to execute migrations: %v", err)
	}

	log.Println("migrations applied")
}
