package repository

import (
	"database/sql"
	"fmt"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"

	"github.com/google/uuid"
)

type OrderFilter struct {
	Status *string
}

type OrderRepository interface {
	GetAllByUserID(userID uuid.UUID, limit, offset int, f OrderFilter) ([]models.Order, int, error)
	GetByID(orderID uuid.UUID) (models.Order, error)
	Create(o models.Order) (models.Order, error)
	CreateItems(orderID uuid.UUID, items []models.OrderItem) ([]models.OrderItem, error)
	UpdateStatus(orderID uuid.UUID, status models.OrderStatus) error
}

type OrderRepo struct {
	db DBTX
}

func NewOrderRepo(db DBTX) OrderRepository {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) GetAllByUserID(userID uuid.UUID, limit, offset int, f OrderFilter) ([]models.Order, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1 AND ($2::text IS NULL OR status = $2)`
	if err := r.db.QueryRow(countQuery, userID, f.Status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting orders: %v", err)
	}

	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		  AND ($2::text IS NULL OR status = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(query, userID, f.Status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying orders: %v", err)
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("error scanning order: %v", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating orders: %v", err)
	}
	return orders, total, nil
}

func (r *OrderRepo) GetByID(orderID uuid.UUID) (models.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1	
	`
	var o models.Order
	err := r.db.QueryRow(query, orderID).
		Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Order{}, error_codes.ErrOrderNotFound
		}
		return models.Order{}, fmt.Errorf("error scanning order: %v", err)
	}

	items, err := r.GetItemsByOrderID(orderID)
	if err != nil {
		return models.Order{}, err
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepo) GetItemsByOrderID(orderID uuid.UUID) ([]models.OrderItem, error) {
	query := `
		SELECT id, order_id, item_id, quantity, price_at_purchase, subtotal
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("error querying order items: %v", err)
	}
	defer rows.Close()

	items := []models.OrderItem{}
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.Subtotal); err != nil {
			return nil, fmt.Errorf("error scanning order item: %v", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %v", err)
	}
	return items, nil
}

func (r *OrderRepo) Create(o models.Order) (models.Order, error) {
	query := `
		INSERT INTO orders (user_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(query, o.UserID, o.Status, o.TotalAmount).
		Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return models.Order{}, fmt.Errorf("error creating order: %v", err)
	}
	return o, nil
}

func (r *OrderRepo) CreateItems(orderID uuid.UUID, items []models.OrderItem) ([]models.OrderItem, error) {
	query := `
		INSERT INTO order_items (order_id, item_id, quantity, price_at_purchase, subtotal)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	for i := range items {
		err := r.db.QueryRow(query, orderID, items[i].ProductID, items[i].Quantity, items[i].UnitPrice, items[i].Subtotal).
			Scan(&items[i].ID)
		if err != nil {
			return nil, fmt.Errorf("error creating order item: %v", err)
		}
		items[i].OrderID = orderID
	}
	return items, nil
}

func (r *OrderRepo) UpdateStatus(orderID uuid.UUID, status models.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(query, status, orderID)
	if err != nil {
		return fmt.Errorf("error updating order status: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if affected == 0 {
		return error_codes.ErrOrderNotFound
	}
	return nil
}
