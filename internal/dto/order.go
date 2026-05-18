package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity  int       `json:"quantity" example:"2"`
}

type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" example:"paid"`
}

type OrderItemResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
}

type OrderResponse struct {
	ID          uuid.UUID           `json:"id"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	Items       []OrderItemResponse `json:"items,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
}
