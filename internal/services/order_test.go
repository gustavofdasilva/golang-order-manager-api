package services_test

import (
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repomocks "golang-order-manager-api/internal/repositories/mocks"
	"golang-order-manager-api/internal/services"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type orderServiceMocks struct {
	txFactory   *repomocks.MockOrderTxFactory
	orderTx     *repomocks.MockOrderTx
	orderRepo   *repomocks.MockOrderRepository
	productRepo *repomocks.MockProductRepository
	cache       *repomocks.MockOrderCache
}

func makeOrderService(t *testing.T) (services.OrderService, *orderServiceMocks) {
	m := &orderServiceMocks{
		txFactory:   repomocks.NewMockOrderTxFactory(t),
		orderTx:     repomocks.NewMockOrderTx(t),
		orderRepo:   repomocks.NewMockOrderRepository(t),
		productRepo: repomocks.NewMockProductRepository(t),
		cache:       repomocks.NewMockOrderCache(t),
	}

	// OrderTx sempre devolve os repos mockados
	m.orderTx.EXPECT().OrderRepository().Return(m.orderRepo).Maybe()
	m.orderTx.EXPECT().ProductRepository().Return(m.productRepo).Maybe()

	svc := services.NewOrderService(m.txFactory, m.orderRepo, m.cache)
	return svc, m
}

func (m *orderServiceMocks) expectTx(t *testing.T) {
	m.txFactory.EXPECT().
		BeginTx(mock.Anything).
		Return(m.orderTx, nil)

	m.orderTx.EXPECT().Rollback().Return(nil).Maybe()
}

func TestGetByID_FromCache(t *testing.T) {
	svc, m := makeOrderService(t)

	orderID := uuid.New()
	cached := models.Order{ID: orderID, Status: models.OrderStatusPending}

	m.cache.EXPECT().
		GetOrder(mock.Anything, orderID).
		Return(cached, nil)

	result, err := svc.GetByID(orderID)

	require.NoError(t, err)
	assert.Equal(t, orderID, result.ID)
	// repo nunca foi chamado
	m.orderRepo.AssertNotCalled(t, "GetByID")
}

func TestGetByID_FromRepo_WhenCacheMiss(t *testing.T) {
	svc, m := makeOrderService(t)

	orderID := uuid.New()
	order := models.Order{ID: orderID, Status: models.OrderStatusPending}

	m.cache.EXPECT().
		GetOrder(mock.Anything, orderID).
		Return(models.Order{}, nil) // cache miss

	m.orderRepo.EXPECT().
		GetByID(orderID).
		Return(order, nil)

	m.cache.EXPECT().
		SetOrder(mock.Anything, order).
		Return(nil)

	result, err := svc.GetByID(orderID)

	require.NoError(t, err)
	assert.Equal(t, orderID, result.ID)
	m.cache.AssertCalled(t, "SetOrder", mock.Anything, order)
}

func TestCreateOrder_Success(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	userID := uuid.New()
	productID := uuid.New()

	product := models.Product{
		ID:    productID,
		Price: 50.0,
		Stock: 10,
	}
	inputs := []services.OrderItemInput{
		{ProductID: productID, Quantity: 2},
	}
	createdOrder := models.Order{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      models.OrderStatusPending,
		TotalAmount: 100.0,
	}
	createdItems := []models.OrderItem{
		{ProductID: productID, Quantity: 2, UnitPrice: 50.0, Subtotal: 100.0},
	}

	m.productRepo.EXPECT().GetByID(productID).Return(product, nil)
	m.orderRepo.EXPECT().
		Create(mock.MatchedBy(func(o models.Order) bool {
			return o.UserID == userID &&
				o.TotalAmount == 100.0 &&
				o.Status == models.OrderStatusPending
		})).
		Return(createdOrder, nil)
	m.orderRepo.EXPECT().
		CreateItems(createdOrder.ID, mock.Anything).
		Return(createdItems, nil)
	m.orderTx.EXPECT().Commit().Return(nil)

	result, err := svc.CreateOrder(userID, inputs)

	require.NoError(t, err)
	assert.Equal(t, 100.0, result.TotalAmount)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, productID, result.Items[0].ProductID)
	assert.Equal(t, 2, result.Items[0].Quantity)
	assert.Equal(t, 50.0, result.Items[0].UnitPrice)
	assert.Equal(t, 100.0, result.Items[0].Subtotal)
	assert.Equal(t, models.OrderStatusPending, result.Status)
}

func TestCreateOrder_EmptyInputs(t *testing.T) {
	svc, _ := makeOrderService(t)

	_, err := svc.CreateOrder(uuid.New(), []services.OrderItemInput{})

	assert.ErrorIs(t, err, error_codes.ErrOrderEmpty)
}

func TestCreateOrder_InvalidQuantity(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	inputs := []services.OrderItemInput{
		{ProductID: uuid.New(), Quantity: 0},
	}

	_, err := svc.CreateOrder(uuid.New(), inputs)

	assert.ErrorIs(t, err, error_codes.ErrInvalidQuantity)
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	productID := uuid.New()
	m.productRepo.EXPECT().
		GetByID(productID).
		Return(models.Product{ID: productID, Stock: 1, Price: 10.0}, nil)

	inputs := []services.OrderItemInput{
		{ProductID: productID, Quantity: 5},
	}

	_, err := svc.CreateOrder(uuid.New(), inputs)

	assert.ErrorIs(t, err, error_codes.ErrInsufficientStock)
}

func TestCreateOrder_ProductNotFound(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	productID := uuid.New()
	m.productRepo.EXPECT().
		GetByID(productID).
		Return(models.Product{}, error_codes.ErrProductNotFound)

	_, err := svc.CreateOrder(uuid.New(), []services.OrderItemInput{
		{ProductID: productID, Quantity: 1},
	})

	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	svc, _ := makeOrderService(t)

	_, err := svc.UpdateStatus(uuid.New(), uuid.New(), "INVALID_STATUS")

	assert.ErrorIs(t, err, error_codes.ErrInvalidOrderStatus)
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	orderID := uuid.New()
	// tries to transition from Paid to Pending, which is invalid
	m.orderRepo.EXPECT().
		GetByID(orderID).
		Return(models.Order{ID: orderID, Status: models.OrderStatusPaid}, nil)

	_, err := svc.UpdateStatus(orderID, uuid.New(), models.OrderStatusPending)

	assert.ErrorIs(t, err, error_codes.ErrOrderStatusTransition)
}

func TestUpdateStatus_ToPaid_DecrementsStock(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	orderID := uuid.New()
	productID := uuid.New()
	order := models.Order{
		ID:     orderID,
		Status: models.OrderStatusPending,
		Items:  []models.OrderItem{{ProductID: productID, Quantity: 3}},
	}

	m.orderRepo.EXPECT().GetByID(orderID).Return(order, nil)
	m.orderRepo.EXPECT().UpdateStatus(orderID, models.OrderStatusPaid).Return(nil)
	m.productRepo.EXPECT().DecrementStock(productID, 3).Return(nil)
	m.cache.EXPECT().DeleteOrder(mock.Anything, orderID).Return(nil)
	m.orderTx.EXPECT().Commit().Return(nil)

	result, err := svc.UpdateStatus(orderID, uuid.New(), models.OrderStatusPaid)

	require.NoError(t, err)
	assert.Equal(t, models.OrderStatusPaid, result.Status)
}

func TestUpdateStatus_ToCancelled_WhenPaid_IncrementsStock(t *testing.T) {
	svc, m := makeOrderService(t)
	m.expectTx(t)

	orderID := uuid.New()
	productID := uuid.New()
	order := models.Order{
		ID:     orderID,
		Status: models.OrderStatusPaid,
		Items:  []models.OrderItem{{ProductID: productID, Quantity: 2}},
	}

	m.orderRepo.EXPECT().GetByID(orderID).Return(order, nil)
	m.orderRepo.EXPECT().UpdateStatus(orderID, models.OrderStatusCancelled).Return(nil)
	m.productRepo.EXPECT().IncrementStock(productID, 2).Return(nil)
	m.cache.EXPECT().DeleteOrder(mock.Anything, orderID).Return(nil)
	m.orderTx.EXPECT().Commit().Return(nil)

	result, err := svc.UpdateStatus(orderID, uuid.New(), models.OrderStatusCancelled)

	require.NoError(t, err)
	assert.Equal(t, models.OrderStatusCancelled, result.Status)
}
