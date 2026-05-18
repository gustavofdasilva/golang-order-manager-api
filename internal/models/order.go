package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusCompleted OrderStatus = "completed"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusPaid, OrderStatusCancelled, OrderStatusCompleted:
		return true
	}
	return false
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	switch s {
	case OrderStatusPending:
		return next == OrderStatusPaid || next == OrderStatusCancelled
	case OrderStatusPaid:
		return next == OrderStatusCompleted || next == OrderStatusCancelled
	}
	return false
}

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Status      OrderStatus
	TotalAmount float64
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
	Subtotal  float64
}
