package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"golang-order-manager-api/internal/dto"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/responses"
	"golang-order-manager-api/internal/services"
	"golang-order-manager-api/pkg/cache"
	"golang-order-manager-api/pkg/database"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func parseOrderFilter(c echo.Context) models.OrderFilter {
	f := models.OrderFilter{}
	if v := c.QueryParam("status"); v != "" {
		f.Status = &v
	}
	return f
}

// ListOrders godoc
//
// @Summary List user orders
// @Description Returns all orders belonging to the authenticated user
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param page   query int    false "Page number" default(1)
// @Param limit  query int    false "Items per page (max 100)" default(10)
// @Param status query string false "Filter by status (pending, paid, completed, cancelled)"
// @Success 200 {object} responses.PaginatedResponse{data=[]dto.OrderResponse}
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders [get]
func ListOrders(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	page, limit := parsePaginationParams(c)
	filter := parseOrderFilter(c)

	db := database.GetDB()

	orderCache := cache.NewRedisOrderCache()

	txFactory := repository.NewTxFactory(db)
	orderRepo := repository.NewOrderRepo(db)

	orderService := services.NewOrderService(txFactory, orderRepo, orderCache)

	orders, total, err := orderService.GetAllByUserID(userID, page, limit, filter)
	if err != nil {
		slog.Error("Failed to list orders", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	result := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		result[i] = toOrderResponse(o)
	}

	totalPages := (total + limit - 1) / limit
	return responses.Paginated(c, http.StatusOK, "Orders retrieved successfully", result, responses.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetOrder godoc
//
// @Summary Get order by ID
// @Description Returns a single order belonging to the authenticated user
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} responses.SuccessResponse{data=dto.OrderResponse}
// @Failure 404 {object} responses.ErrorResponse "Order not found"
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders/{id} [get]
func GetOrder(c echo.Context) error {

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, errors.New("invalid order id"))
	}

	db := database.GetDB()

	orderCache := cache.NewRedisOrderCache()

	txFactory := repository.NewTxFactory(db)
	orderRepo := repository.NewOrderRepo(db)

	orderService := services.NewOrderService(txFactory, orderRepo, orderCache)

	order, err := orderService.GetByID(orderID)
	if err != nil {
		if errors.Is(err, error_codes.ErrOrderNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		slog.Error("Failed to get order", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return responses.Success(c, http.StatusOK, "Order retrieved successfully", toOrderResponse(order))
}

// CreateOrder godoc
//
// @Summary Create an order
// @Description Creates a new order with the provided items. Total is calculated by the server using prices frozen at purchase time.
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order payload"
// @Success 201 {object} responses.SuccessResponse{data=dto.OrderResponse}
// @Failure 400 {object} responses.ErrorResponse "Empty order, invalid quantity, or insufficient stock"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders [post]
func CreateOrder(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	req := dto.CreateOrderRequest{}
	c.Bind(&req)

	//TODO? Why the orderItemInput struct in services? Why not in dto?
	inputs := make([]services.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		inputs[i] = services.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	db := database.GetDB()

	orderCache := cache.NewRedisOrderCache()

	txFactory := repository.NewTxFactory(db)
	orderRepo := repository.NewOrderRepo(db)

	orderService := services.NewOrderService(txFactory, orderRepo, orderCache)

	order, err := orderService.CreateOrder(userID, inputs)
	if err != nil {
		switch {
		case errors.Is(err, error_codes.ErrOrderEmpty):
			return responses.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, error_codes.ErrInvalidQuantity):
			return responses.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, error_codes.ErrInsufficientStock):
			return responses.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, error_codes.ErrProductNotFound):
			return responses.Error(c, http.StatusNotFound, err)
		default:
			slog.Error("Failed to create order", slog.Any("err", err))
			return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
		}
	}

	return responses.Success(c, http.StatusCreated, "Order created successfully", toOrderResponse(order))
}

// UpdateOrderStatus godoc
//
// @Summary Update order status
// @Description Transitions the order status. Valid transitions: pending→paid, pending→cancelled, paid→completed, paid→cancelled.
// @Description Paying an order decrements product stock; cancelling a paid order restores it.
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.UpdateOrderStatusRequest true "New status"
// @Success 200 {object} responses.SuccessResponse{data=dto.OrderResponse}
// @Failure 400 {object} responses.ErrorResponse "Invalid status or illegal transition"
// @Failure 404 {object} responses.ErrorResponse "Order not found"
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders/{id}/status [patch]
func UpdateOrderStatus(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, errors.New("invalid order id"))
	}

	req := dto.UpdateOrderStatusRequest{}
	c.Bind(&req)

	newStatus := models.OrderStatus(req.Status)

	db := database.GetDB()

	orderCache := cache.NewRedisOrderCache()

	txFactory := repository.NewTxFactory(db)
	orderRepo := repository.NewOrderRepo(db)

	orderService := services.NewOrderService(txFactory, orderRepo, orderCache)

	order, err := orderService.UpdateStatus(orderID, userID, newStatus)
	if err != nil {
		switch {
		case errors.Is(err, error_codes.ErrInvalidOrderStatus):
			return responses.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, error_codes.ErrOrderStatusTransition):
			return responses.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, error_codes.ErrInsufficientStock):
			return responses.Error(c, http.StatusBadRequest, err)
		case errors.Is(err, error_codes.ErrOrderNotFound):
			return responses.Error(c, http.StatusNotFound, err)
		default:
			slog.Error("Failed to update order status", slog.Any("err", err))
			return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
		}
	}

	return responses.Success(c, http.StatusOK, "Order status updated successfully", toOrderResponse(order))
}

func toOrderResponse(o models.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = dto.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
		}
	}
	return dto.OrderResponse{
		ID:          o.ID,
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		Items:       items,
		CreatedAt:   o.CreatedAt,
	}
}
