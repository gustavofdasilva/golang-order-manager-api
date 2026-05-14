package services

import (
	"database/sql"
	"fmt"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
}

// TODO? db should be a dependency of the repository, not the service. The service should receive the repository as a dependency, not the db. This way we can mock the repository in tests and avoid using a real database connection.
type OrderService struct {
	db *sql.DB
}

func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) GetAllByUserID(userID uuid.UUID) ([]models.Order, error) {
	repo := repository.NewOrderRepo(s.db)
	return repo.GetAllByUserID(userID)
}

func (s *OrderService) GetByID(orderID, userID uuid.UUID) (models.Order, error) {
	repo := repository.NewOrderRepo(s.db)
	return repo.GetByIDAndUserID(orderID, userID)
}

func (s *OrderService) CreateOrder(userID uuid.UUID, inputs []OrderItemInput) (models.Order, error) {
	if len(inputs) == 0 {
		return models.Order{}, error_codes.ErrOrderEmpty
	}

	tx, err := s.db.Begin()
	if err != nil {
		return models.Order{}, fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	productRepo := repository.NewProductRepo(tx)
	orderRepo := repository.NewOrderRepo(tx)

	var orderItems []models.OrderItem
	total := 0.0

	for _, input := range inputs {
		if input.Quantity <= 0 {
			return models.Order{}, error_codes.ErrInvalidQuantity
		}

		product, err := productRepo.GetByID(input.ProductID)
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

	order, err := orderRepo.Create(models.Order{
		UserID:      userID,
		Status:      models.OrderStatusPending,
		TotalAmount: total,
	})
	if err != nil {
		return models.Order{}, err
	}

	createdItems, err := orderRepo.CreateItems(order.ID, orderItems)
	if err != nil {
		return models.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Order{}, fmt.Errorf("error committing transaction: %v", err)
	}

	order.Items = createdItems
	return order, nil
}

// TODO: Refactor this method to reduce its Cognitive Complexity from 19 to the 15 allowed. [+13 locations]
func (s *OrderService) UpdateStatus(orderID, userID uuid.UUID, newStatus models.OrderStatus) (models.Order, error) {
	if !newStatus.IsValid() {
		return models.Order{}, error_codes.ErrInvalidOrderStatus
	}

	tx, err := s.db.Begin()
	if err != nil {
		return models.Order{}, fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	orderRepo := repository.NewOrderRepo(tx)
	productRepo := repository.NewProductRepo(tx)

	order, err := orderRepo.GetByIDAndUserID(orderID, userID)
	if err != nil {
		return models.Order{}, err
	}

	if !order.Status.CanTransitionTo(newStatus) {
		return models.Order{}, error_codes.ErrOrderStatusTransition
	}

	// Decrement stock when order is paid
	if newStatus == models.OrderStatusPaid {
		for _, item := range order.Items {
			if err := productRepo.DecrementStock(item.ProductID, item.Quantity); err != nil {
				return models.Order{}, err
			}
		}
	}

	// Restore stock when a paid order is cancelled
	if newStatus == models.OrderStatusCancelled && order.Status == models.OrderStatusPaid {
		for _, item := range order.Items {
			if err := productRepo.IncrementStock(item.ProductID, item.Quantity); err != nil {
				return models.Order{}, err
			}
		}
	}

	if err := orderRepo.UpdateStatus(orderID, newStatus); err != nil {
		return models.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Order{}, fmt.Errorf("error committing transaction: %v", err)
	}

	order.Status = newStatus
	return order, nil
}
