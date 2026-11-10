package repository

import (
	"context"
	"database/sql"
)

type sqlOrderTx struct {
	tx          *sql.Tx
	orderRepo   OrderRepository
	productRepo ProductRepository
}

func (t *sqlOrderTx) OrderRepository() OrderRepository     { return t.orderRepo }
func (t *sqlOrderTx) ProductRepository() ProductRepository { return t.productRepo }
func (t *sqlOrderTx) Commit() error                        { return t.tx.Commit() }
func (t *sqlOrderTx) Rollback() error                      { return t.tx.Rollback() }

type SqlTxFactory struct {
	db *sql.DB
}

func NewTxFactory(db *sql.DB) OrderTxFactory {
	return &SqlTxFactory{db: db}
}

func (f *SqlTxFactory) BeginTx(ctx context.Context) (OrderTx, error) {
	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqlOrderTx{
		tx:          tx,
		orderRepo:   NewOrderRepo(tx),
		productRepo: NewProductRepo(tx),
	}, nil
}
