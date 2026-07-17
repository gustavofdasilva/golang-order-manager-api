package testhelpers

import (
	"database/sql"
	"testing"

	"golang-order-manager-api/internal/models"

	"github.com/google/uuid"
)

func CreateUserFixture(t *testing.T, db *sql.DB) models.User {
	t.Helper()
	user := models.User{
		Username: "testuser_" + uuid.New().String()[:8],
		Email:    "test_" + uuid.New().String()[:8] + "@example.com",
		Password: "hashed-password",
	}
	err := db.QueryRow(
		`INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`,
		user.Username, user.Email, user.Password,
	).Scan(&user.ID)
	if err != nil {
		t.Fatalf("failed to create user fixture: %v", err)
	}
	return user
}

func CreateProductFixture(t *testing.T, db *sql.DB, price float64, stock int) models.Product {
	t.Helper()
	product := models.Product{
		Name:  "product_" + uuid.New().String()[:8],
		Price: price,
		Stock: stock,
	}
	err := db.QueryRow(
		`INSERT INTO items (name, price, stock) VALUES ($1, $2, $3) RETURNING id, created_at`,
		product.Name, product.Price, product.Stock,
	).Scan(&product.ID, &product.CreatedAt)
	if err != nil {
		t.Fatalf("failed to create product fixture: %v", err)
	}
	return product
}

func CreateOrderFixture(t *testing.T, db *sql.DB, userID *uuid.UUID) models.Order {
	t.Helper()

	var orderUserID uuid.UUID

	if userID == nil {
		user := CreateUserFixture(t, db)
		orderUserID = user.ID
	} else {
		orderUserID = *userID
	}

	productA := CreateProductFixture(t, db, 10, 10)
	productB := CreateProductFixture(t, db, 10, 10)

	orderItemA := models.OrderItem{
		ProductID: productA.ID,
		Quantity:  5,
		UnitPrice: 10,
		Subtotal:  50,
	}

	orderItemB := models.OrderItem{
		ProductID: productB.ID,
		Quantity:  5,
		UnitPrice: 10,
		Subtotal:  50,
	}

	order := models.Order{
		UserID:      orderUserID,
		Status:      models.OrderStatusPending,
		TotalAmount: 100.0,
		Items: []models.OrderItem{
			orderItemA, orderItemB,
		},
	}

	err := db.QueryRow(
		`INSERT INTO orders (user_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`,
		orderUserID, order.Status, order.TotalAmount,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to create order fixture: %v", err)
	}

	query := `
		INSERT INTO order_items (order_id, item_id, quantity, price_at_purchase, subtotal)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	for i := range order.Items {
		err := db.QueryRow(query, order.ID, order.Items[i].ProductID, order.Items[i].Quantity, order.Items[i].UnitPrice, order.Items[i].Subtotal).
			Scan(&order.Items[i].ID)
		if err != nil {
			t.Fatalf("failed to create order item fixture: %v", err)
		}
		order.Items[i].OrderID = order.ID
	}

	return order
}
