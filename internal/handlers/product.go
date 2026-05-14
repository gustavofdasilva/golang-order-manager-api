package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"golang-order-manager-api/internal/dto"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/responses"
	"golang-order-manager-api/internal/services"
	"golang-order-manager-api/pkg/database"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const errInvalidProductID = "invalid product id"

func parsePaginationParams(c echo.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ = strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return
}

// ListProducts godoc
//
// @Summary List all products
// @Description Returns all active products
// @Tags Products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 100)" default(10)
// @Success 200 {object} responses.PaginatedResponse{data=[]dto.ProductResponse}
// @Failure 500 {object} responses.ErrorResponse
// @Router /products [get]
func ListProducts(c echo.Context) error {
	page, limit := parsePaginationParams(c)

	db := database.GetDB()
	repo := repository.NewProductRepo(db)
	productService := services.NewProductService(&repo)

	products, total, err := productService.GetAll(page, limit)
	if err != nil {
		slog.Error("Failed to list products", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	result := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		result[i] = dto.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CreatedAt:   p.CreatedAt,
		}
	}

	totalPages := (total + limit - 1) / limit
	return responses.Paginated(c, http.StatusOK, "Products retrieved successfully", result, responses.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetProduct godoc
//
// @Summary Get product by ID
// @Description Returns a single active product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} responses.SuccessResponse{data=dto.ProductResponse}
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse
// @Router /products/{id} [get]
func GetProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, errors.New(errInvalidProductID))
	}

	db := database.GetDB()
	repo := repository.NewProductRepo(db)
	productService := services.NewProductService(&repo)

	product, err := productService.GetByID(id)
	if err != nil {
		if errors.Is(err, error_codes.ErrProductNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}
		slog.Error("Failed to get product", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return responses.Success(c, http.StatusOK, "Product retrieved successfully", dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
	})
}

// CreateProduct godoc
//
// @Summary Create a product
// @Description Creates a new product
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateProductRequest true "Product payload"
// @Success 201 {object} responses.SuccessResponse{data=dto.ProductResponse}
// @Failure 400 {object} responses.ErrorResponse "Invalid fields"
// @Failure 500 {object} responses.ErrorResponse
// @Router /products [post]
func CreateProduct(c echo.Context) error {
	req := dto.CreateProductRequest{}
	c.Bind(&req)

	if req.Name == "" {
		return responses.Error(c, http.StatusBadRequest, errors.New("name is required"))
	}

	db := database.GetDB()
	repo := repository.NewProductRepo(db)
	productService := services.NewProductService(&repo)

	product, err := productService.Create(models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		if errors.Is(err, error_codes.ErrInvalidPrice) || errors.Is(err, error_codes.ErrInvalidStock) {
			return responses.Error(c, http.StatusBadRequest, err)
		}
		slog.Error("Failed to create product", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return responses.Success(c, http.StatusCreated, "Product created successfully", dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
	})
}

// UpdateProduct godoc
//
// @Summary Update a product
// @Description Updates an existing product's fields
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body dto.UpdateProductRequest true "Update payload"
// @Success 200 {object} responses.SuccessResponse{data=dto.ProductResponse}
// @Failure 400 {object} responses.ErrorResponse "Invalid fields"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse
// @Router /products/{id} [patch]
func UpdateProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, errors.New(errInvalidProductID))
	}

	req := dto.UpdateProductRequest{}
	c.Bind(&req)

	db := database.GetDB()
	repo := repository.NewProductRepo(db)
	productService := services.NewProductService(&repo)

	product, err := productService.Update(id, req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		if errors.Is(err, error_codes.ErrProductNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}
		if errors.Is(err, error_codes.ErrInvalidPrice) || errors.Is(err, error_codes.ErrInvalidStock) {
			return responses.Error(c, http.StatusBadRequest, err)
		}
		slog.Error("Failed to update product", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return responses.Success(c, http.StatusOK, "Product updated successfully", dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
	})
}

// DeleteProduct godoc
//
// @Summary Delete a product
// @Description Soft-deletes a product
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 "Product deleted successfully"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse
// @Router /products/{id} [delete]
func DeleteProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, errors.New(errInvalidProductID))
	}

	db := database.GetDB()
	repo := repository.NewProductRepo(db)
	productService := services.NewProductService(&repo)

	if err := productService.Delete(id); err != nil {
		if errors.Is(err, error_codes.ErrProductNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}
		slog.Error("Failed to delete product", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return c.NoContent(http.StatusNoContent)
}
