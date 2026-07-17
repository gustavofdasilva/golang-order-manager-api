package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/models"
	"golang-order-manager-api/internal/services"
	svcmocks "golang-order-manager-api/internal/services/mocks"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeOrderHandler(t *testing.T) (*handlers.OrderHandler, *svcmocks.MockOrderService) {
	mockSvc := svcmocks.NewMockOrderService(t)
	h := handlers.NewOrderHandler(mockSvc)
	return h, mockSvc
}

func newOrderContext(method, path, body string, userID uuid.UUID, pathParams map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", userID)

	if pathParams != nil {
		names := make([]string, 0, len(pathParams))
		values := make([]string, 0, len(pathParams))
		for k, v := range pathParams {
			names = append(names, k)
			values = append(values, v)
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
	}

	return c, rec
}

func fakeOrder(userID uuid.UUID) models.Order {
	productID := uuid.New()
	return models.Order{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      models.OrderStatusPending,
		TotalAmount: 100.0,
		CreatedAt:   time.Now(),
		Items: []models.OrderItem{
			{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  2,
				UnitPrice: 50.0,
				Subtotal:  100.0,
			},
		},
	}
}

func TestListOrders_Success(t *testing.T) {
	h, m := makeOrderHandler(t)

	userID := uuid.New()
	orders := []models.Order{fakeOrder(userID), fakeOrder(userID)}

	m.EXPECT().
		GetAllByUserID(userID, 1, 10, mock.Anything).
		Return(orders, 2, nil)

	c, rec := newOrderContext(http.MethodGet, "/orders", "", userID, nil)
	err := h.ListOrders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := parseBody(t, rec)
	data := body["data"].([]any)
	assert.Len(t, data, 2)

	pagination := body["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(1), pagination["total_pages"])
}

func TestListOrders_WithStatusFilter(t *testing.T) {
	h, m := makeOrderHandler(t)

	userID := uuid.New()
	m.EXPECT().
		GetAllByUserID(userID, 1, 10, mock.Anything).
		Return([]models.Order{fakeOrder(userID)}, 1, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/orders?status=pending", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", userID)

	err := h.ListOrders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListOrders_InternalError(t *testing.T) {
	h, m := makeOrderHandler(t)

	userID := uuid.New()
	m.EXPECT().
		GetAllByUserID(userID, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, 0, fmt.Errorf("db error"))

	c, rec := newOrderContext(http.MethodGet, "/orders", "", userID, nil)
	err := h.ListOrders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- GetOrder ---

func TestGetOrder_Success(t *testing.T) {
	h, m := makeOrderHandler(t)

	userID := uuid.New()
	order := fakeOrder(userID)

	m.EXPECT().GetByID(order.ID).Return(order, nil)

	c, rec := newOrderContext(http.MethodGet, "/orders/"+order.ID.String(), "", userID, map[string]string{"id": order.ID.String()})
	err := h.GetOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, order.ID.String(), data["id"])
	assert.Equal(t, "pending", data["status"])
	assert.Equal(t, 100.0, data["total_amount"])

	items := data["items"].([]any)
	assert.Len(t, items, 1)
}

func TestGetOrder_NotFound(t *testing.T) {
	h, m := makeOrderHandler(t)

	orderID := uuid.New()
	m.EXPECT().GetByID(orderID).Return(models.Order{}, error_codes.ErrOrderNotFound)

	c, rec := newOrderContext(http.MethodGet, "/orders/"+orderID.String(), "", uuid.New(), map[string]string{"id": orderID.String()})
	err := h.GetOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetOrder_InvalidID(t *testing.T) {
	h, _ := makeOrderHandler(t)

	c, rec := newOrderContext(http.MethodGet, "/orders/not-a-uuid", "", uuid.New(), map[string]string{"id": "not-a-uuid"})
	err := h.GetOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- CreateOrder ---

func TestCreateOrder_Success(t *testing.T) {
	h, m := makeOrderHandler(t)

	userID := uuid.New()
	productID := uuid.New()
	order := fakeOrder(userID)

	body := fmt.Sprintf(`{"items":[{"product_id":"%s","quantity":2}]}`, productID)

	m.EXPECT().
		CreateOrder(userID, []services.OrderItemInput{
			{ProductID: productID, Quantity: 2},
		}).
		Return(order, nil)

	c, rec := newOrderContext(http.MethodPost, "/orders", body, userID, nil)
	err := h.CreateOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, "pending", data["status"])
	assert.Equal(t, 100.0, data["total_amount"])
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	h, m := makeOrderHandler(t)

	m.EXPECT().
		CreateOrder(mock.Anything, []services.OrderItemInput{}).
		Return(models.Order{}, error_codes.ErrOrderEmpty)

	c, rec := newOrderContext(http.MethodPost, "/orders", `{"items":[]}`, uuid.New(), nil)
	err := h.CreateOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_InvalidQuantity(t *testing.T) {
	h, m := makeOrderHandler(t)

	productID := uuid.New()
	m.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(models.Order{}, error_codes.ErrInvalidQuantity)

	body := fmt.Sprintf(`{"items":[{"product_id":"%s","quantity":0}]}`, productID)
	c, rec := newOrderContext(http.MethodPost, "/orders", body, uuid.New(), nil)
	err := h.CreateOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	h, m := makeOrderHandler(t)

	m.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(models.Order{}, error_codes.ErrInsufficientStock)

	body := fmt.Sprintf(`{"items":[{"product_id":"%s","quantity":999}]}`, uuid.New())
	c, rec := newOrderContext(http.MethodPost, "/orders", body, uuid.New(), nil)
	err := h.CreateOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_ProductNotFound(t *testing.T) {
	h, m := makeOrderHandler(t)

	m.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(models.Order{}, error_codes.ErrProductNotFound)

	body := fmt.Sprintf(`{"items":[{"product_id":"%s","quantity":1}]}`, uuid.New())
	c, rec := newOrderContext(http.MethodPost, "/orders", body, uuid.New(), nil)
	err := h.CreateOrder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	h, m := makeOrderHandler(t)

	userID := uuid.New()
	order := fakeOrder(userID)
	order.Status = models.OrderStatusPaid

	m.EXPECT().
		UpdateStatus(order.ID, userID, models.OrderStatusPaid).
		Return(order, nil)

	body := `{"status":"paid"}`
	c, rec := newOrderContext(http.MethodPatch, "/orders/"+order.ID.String()+"/status", body, userID, map[string]string{"id": order.ID.String()})
	err := h.UpdateOrderStatus(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, "paid", data["status"])
}

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	h, m := makeOrderHandler(t)

	orderID := uuid.New()
	userID := uuid.New()

	m.EXPECT().
		UpdateStatus(orderID, userID, models.OrderStatus("invalid")).
		Return(models.Order{}, error_codes.ErrInvalidOrderStatus)

	body := `{"status":"invalid"}`
	c, rec := newOrderContext(http.MethodPatch, "/orders/"+orderID.String()+"/status", body, userID, map[string]string{"id": orderID.String()})
	err := h.UpdateOrderStatus(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateOrderStatus_InvalidTransition(t *testing.T) {
	h, m := makeOrderHandler(t)

	orderID := uuid.New()
	userID := uuid.New()

	m.EXPECT().
		UpdateStatus(orderID, userID, models.OrderStatusPending).
		Return(models.Order{}, error_codes.ErrOrderStatusTransition)

	body := `{"status":"pending"}`
	c, rec := newOrderContext(http.MethodPatch, "/orders/"+orderID.String()+"/status", body, userID, map[string]string{"id": orderID.String()})
	err := h.UpdateOrderStatus(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateOrderStatus_OrderNotFound(t *testing.T) {
	h, mockSvc := makeOrderHandler(t)

	orderID := uuid.New()
	userID := uuid.New()

	mockSvc.EXPECT().
		UpdateStatus(orderID, userID, models.OrderStatusPaid).
		Return(models.Order{}, error_codes.ErrOrderNotFound)

	body := `{"status":"paid"}`
	c, rec := newOrderContext(http.MethodPatch, "/orders/"+orderID.String()+"/status", body, userID, map[string]string{"id": orderID.String()})
	err := h.UpdateOrderStatus(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateOrderStatus_InvalidID(t *testing.T) {
	h, _ := makeOrderHandler(t)

	c, rec := newOrderContext(http.MethodPatch, "/orders/bad-id/status", `{"status":"paid"}`, uuid.New(), map[string]string{"id": "bad-id"})
	err := h.UpdateOrderStatus(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
