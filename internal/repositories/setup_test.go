package repository_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"golang-order-manager-api/pkg/database"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer container.Terminate(ctx)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic("failed to open db: " + err.Error())
	}
	defer testDB.Close()

	if err := testDB.Ping(); err != nil {
		panic("failed to ping db: " + err.Error())
	}

	database.RunMigrations(testDB)

	os.Exit(m.Run())
}

func ClearDatabase(db *sql.DB) {
	_, err := db.Exec(`
		DO $$ 
		DECLARE 
			r RECORD;
		BEGIN 
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP 
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE;';
			END LOOP; 
		END $$;
	`)
	if err != nil {
		log.Fatalf("error to clear database: %v", err)
	}
}
