package services

import (
	"context"
	"fmt"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"

	"github.com/google/uuid"
)

type OrderCache interface {
	GetOrder(ctx context.Context, id uuid.UUID) (models.Order, error)
	SetOrder(ctx context.Context, order models.Order) error
	DeleteOrder(ctx context.Context, id uuid.UUID) error
}

type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
}

type OrderService struct {
	txFactory repository.OrderTxFactory
	repo      repository.OrderRepository
	cache     OrderCache
}

func NewOrderService(txFactory repository.OrderTxFactory, repo repository.OrderRepository, cache OrderCache) *OrderService {
	return &OrderService{txFactory: txFactory, repo: repo, cache: cache}
}

func (s *OrderService) GetAllByUserID(userID uuid.UUID, page, limit int, filter repository.OrderFilter) ([]models.Order, int, error) {
	offset := (page - 1) * limit
	return s.repo.GetAllByUserID(userID, limit, offset, filter)
}

func (s *OrderService) GetByID(orderID uuid.UUID) (models.Order, error) {
	order, err := s.cache.GetOrder(context.Background(), orderID)
	if err == nil && order.ID != uuid.Nil {
		return order, nil
	}

	order, err = s.repo.GetByID(orderID)

	s.cache.SetOrder(context.Background(), order)

	return order, err
}

func (s *OrderService) CreateOrder(userID uuid.UUID, inputs []OrderItemInput) (models.Order, error) {
	if len(inputs) == 0 {
		return models.Order{}, error_codes.ErrOrderEmpty
	}

	tx, err := s.txFactory.BeginTx(context.Background())
	if err != nil {
		return models.Order{}, fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	var orderItems []models.OrderItem
	total := 0.0

	for _, input := range inputs {
		if input.Quantity <= 0 {
			return models.Order{}, error_codes.ErrInvalidQuantity
		}

		product, err := tx.ProductRepository().GetByID(input.ProductID)
		if err != nil {
			return models.Order{}, err
		}

		if product.Stock < input.Quantity {
			return models.Order{}, error_codes.ErrInsufficientStock
		}

		subtotal := float64(input.Quantity) * product.Price
		total += subtotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
			UnitPrice: product.Price,
			Subtotal:  subtotal,
		})
	}

	order, err := tx.OrderRepository().Create(models.Order{
		UserID:      userID,
		Status:      models.OrderStatusPending,
		TotalAmount: total,
	})
	if err != nil {
		return models.Order{}, err
	}

	createdItems, err := tx.OrderRepository().CreateItems(order.ID, orderItems)
	if err != nil {
		return models.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Order{}, fmt.Errorf("error committing transaction: %v", err)
	}

	order.Items = createdItems
	return order, nil
}

func (s *OrderService) UpdateStatus(orderID, userID uuid.UUID, newStatus models.OrderStatus) (models.Order, error) {
	if !newStatus.IsValid() {
		return models.Order{}, error_codes.ErrInvalidOrderStatus
	}

	tx, err := s.txFactory.BeginTx(context.Background())
	if err != nil {
		return models.Order{}, fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	order, err := tx.OrderRepository().GetByID(orderID)
	if err != nil {
		return models.Order{}, err
	}

	if !order.Status.CanTransitionTo(newStatus) {
		return models.Order{}, error_codes.ErrOrderStatusTransition
	}

	if newStatus == models.OrderStatusPaid {
		if err := decrementStockForItems(tx.ProductRepository(), order.Items); err != nil {
			return models.Order{}, err
		}
	}

	if newStatus == models.OrderStatusCancelled && order.Status == models.OrderStatusPaid {
		if err := incrementStockForItems(tx.ProductRepository(), order.Items); err != nil {
			return models.Order{}, err
		}
	}

	if err := tx.OrderRepository().UpdateStatus(orderID, newStatus); err != nil {
		return models.Order{}, err
	}

	s.cache.DeleteOrder(context.Background(), orderID)

	if err := tx.Commit(); err != nil {
		return models.Order{}, fmt.Errorf("error committing transaction: %v", err)
	}

	order.Status = newStatus
	return order, nil
}

// TODO: Decrementing stock for items should be done in a more efficient way, possibly in a single query or using batch updates. This is a simple implementation for demonstration purposes.
func decrementStockForItems(productRepo repository.ProductRepository, items []models.OrderItem) error {
	for _, item := range items {
		if err := productRepo.DecrementStock(item.ProductID, item.Quantity); err != nil {
			return err
		}
	}
	return nil
}

// TODO: Increment stock for items should be done in a more efficient way, possibly in a single query or using batch updates. This is a simple implementation for demonstration purposes.
func incrementStockForItems(productRepo repository.ProductRepository, items []models.OrderItem) error {
	for _, item := range items {
		if err := productRepo.IncrementStock(item.ProductID, item.Quantity); err != nil {
			return err
		}
	}
	return nil
}
