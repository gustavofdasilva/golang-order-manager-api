package repository

import (
	"context"
	"database/sql"
)

// DBTX is satisfied by both *sql.DB and *sql.Tx, enabling transactional repos.
type DBTX interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type OrderTx interface {
	OrderRepository() OrderRepository
	ProductRepository() ProductRepository
	Commit() error
	Rollback() error
}

type OrderTxFactory interface {
	BeginTx(ctx context.Context) (OrderTx, error)
}
