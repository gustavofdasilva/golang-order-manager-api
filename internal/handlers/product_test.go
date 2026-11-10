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
	svcmocks "golang-order-manager-api/internal/services/mocks"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeProductHandler(t *testing.T) (*handlers.ProductHandler, *svcmocks.MockProductService) {
	m := svcmocks.NewMockProductService(t)
	h := handlers.NewProductHandler(m)
	return h, m
}

func newProductContext(method, path, body string, pathParams map[string]string) (echo.Context, *httptest.ResponseRecorder) {
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

func fakeProduct() models.Product {
	return models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.90,
		Stock:       10,
		CreatedAt:   time.Now(),
	}
}

func TestListProducts_Success(t *testing.T) {
	h, m := makeProductHandler(t)

	products := []models.Product{fakeProduct(), fakeProduct()}
	m.EXPECT().
		GetAll(1, 10, mock.Anything).
		Return(products, 2, nil)

	c, rec := newProductContext(http.MethodGet, "/products", "", nil)
	err := h.ListProducts(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := parseJSON(t, rec)
	data := body["data"].([]any)
	assert.Len(t, data, 2)

	pagination := body["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(1), pagination["total_pages"])
}

func TestListProducts_WithFilters(t *testing.T) {
	h, m := makeProductHandler(t)

	m.EXPECT().
		GetAll(2, 5, mock.Anything).
		Return([]models.Product{fakeProduct()}, 1, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/products?page=2&limit=5&name=test&in_stock=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListProducts(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListProducts_InternalError(t *testing.T) {
	h, m := makeProductHandler(t)

	m.EXPECT().
		GetAll(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, 0, fmt.Errorf("db error"))

	c, rec := newProductContext(http.MethodGet, "/products", "", nil)
	err := h.ListProducts(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetProductByID_Success(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	m.EXPECT().GetByID(product.ID).Return(product, nil)

	c, rec := newProductContext(http.MethodGet, "/products/"+product.ID.String(), "", map[string]string{"id": product.ID.String()})

	err := h.GetProduct(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseJSON(t, rec)["data"].(map[string]any)
	assert.Equal(t, product.ID.String(), data["id"])
	assert.Equal(t, product.Name, data["name"])
	assert.Equal(t, product.Description, data["description"])
	assert.Equal(t, product.Price, data["price"])
	assert.Equal(t, float64(product.Stock), data["stock"])
}

func TestGetProductByID_NotFound(t *testing.T) {
	h, m := makeProductHandler(t)

	productID := uuid.New()

	m.EXPECT().GetByID(productID).Return(models.Product{}, error_codes.ErrProductNotFound)

	c, rec := newProductContext(http.MethodGet, "/products/"+productID.String(), "", map[string]string{"id": productID.String()})

	err := h.GetProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetProduct_InvalidID(t *testing.T) {
	h, _ := makeProductHandler(t)

	c, rec := newProductContext(http.MethodGet, "/products/not-a-uuid", "", map[string]string{"id": "not-a-uuid"})
	err := h.GetProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_Success(t *testing.T) {
	h, m := makeProductHandler(t)

	productJSON := toJSON(t, fakeProduct())

	m.EXPECT().Create(mock.AnythingOfType("models.Product")).Return(fakeProduct(), nil)

	c, rec := newProductContext(http.MethodPost, "/products", productJSON, nil)

	err := h.CreateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	data := parseJSON(t, rec)["data"].(map[string]any)
	assert.Equal(t, data["name"], fakeProduct().Name)
	assert.Equal(t, data["description"], fakeProduct().Description)
}

func TestCreateProduct_InvalidStock(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	product.Stock = -5

	productJSON := toJSON(t, product)

	m.EXPECT().Create(mock.AnythingOfType("models.Product")).Return(models.Product{}, error_codes.ErrInvalidStock)

	c, rec := newProductContext(http.MethodPost, "/products", productJSON, nil)

	err := h.CreateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_InvalidPrice(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	product.Price = -5

	productJSON := toJSON(t, product)

	m.EXPECT().Create(mock.AnythingOfType("models.Product")).Return(models.Product{}, error_codes.ErrInvalidPrice)

	c, rec := newProductContext(http.MethodPost, "/products", productJSON, nil)

	err := h.CreateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_MissingName(t *testing.T) {
	h, _ := makeProductHandler(t)

	product := fakeProduct()

	product.Name = ""

	productJSON := toJSON(t, product)

	c, rec := newProductContext(http.MethodPost, "/products", productJSON, nil)

	err := h.CreateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_InternalError(t *testing.T) {
	h, m := makeProductHandler(t)

	productJSON := toJSON(t, fakeProduct())

	m.EXPECT().Create(mock.AnythingOfType("models.Product")).Return(models.Product{}, fmt.Errorf("db lost connection"))

	c, rec := newProductContext(http.MethodPost, "/products", productJSON, nil)

	err := h.CreateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateProduct_Success(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	productJSON := toJSON(t, product)

	m.EXPECT().Update(product.ID, product.Name, product.Description, &product.Price, &product.Stock).Return(models.Product{ID: product.ID, Name: product.Name, Description: product.Description, Price: product.Price, Stock: product.Stock}, nil)

	c, rec := newProductContext(http.MethodPut, "/products/"+product.ID.String(), productJSON, map[string]string{"id": product.ID.String()})

	err := h.UpdateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, product.ID.String(), data["id"])
	assert.Equal(t, product.Name, data["name"])
	assert.Equal(t, product.Description, data["description"])
	assert.Equal(t, product.Price, data["price"])
	assert.Equal(t, float64(product.Stock), data["stock"])
}

func TestUpdateProduct_NotFound(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	productJSON := toJSON(t, product)

	m.EXPECT().Update(product.ID, product.Name, product.Description, &product.Price, &product.Stock).Return(models.Product{}, error_codes.ErrProductNotFound)

	c, rec := newProductContext(http.MethodPut, "/products/"+product.ID.String(), productJSON, map[string]string{"id": product.ID.String()})

	err := h.UpdateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateProduct_InvalidStock(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	product.Stock = -5

	productJSON := toJSON(t, product)

	m.EXPECT().Update(product.ID, product.Name, product.Description, &product.Price, &product.Stock).Return(models.Product{}, error_codes.ErrInvalidStock)

	c, rec := newProductContext(http.MethodPut, "/products/"+product.ID.String(), productJSON, map[string]string{"id": product.ID.String()})

	err := h.UpdateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateProduct_InvalidPrice(t *testing.T) {
	h, m := makeProductHandler(t)

	product := fakeProduct()

	product.Price = -10.0

	productJSON := toJSON(t, product)

	m.EXPECT().Update(product.ID, product.Name, product.Description, &product.Price, &product.Stock).Return(models.Product{}, error_codes.ErrInvalidPrice)

	c, rec := newProductContext(http.MethodPut, "/products/"+product.ID.String(), productJSON, map[string]string{"id": product.ID.String()})

	err := h.UpdateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateProduct_InvalidID(t *testing.T) {
	h, _ := makeProductHandler(t)

	product := fakeProduct()

	productJSON := toJSON(t, product)

	c, rec := newProductContext(http.MethodPut, "/products/not-a-uuid", productJSON, map[string]string{"id": "not-a-uuid"})

	err := h.UpdateProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteProduct_Success(t *testing.T) {
	h, m := makeProductHandler(t)

	id := uuid.New()
	m.EXPECT().Delete(id).Return(nil)

	c, rec := newProductContext(http.MethodDelete, "/products/"+id.String(), "", map[string]string{"id": id.String()})
	err := h.DeleteProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteProduct_NotFound(t *testing.T) {
	h, m := makeProductHandler(t)

	id := uuid.New()
	m.EXPECT().Delete(id).Return(error_codes.ErrProductNotFound)

	c, rec := newProductContext(http.MethodDelete, "/products/"+id.String(), "", map[string]string{"id": id.String()})
	err := h.DeleteProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteProduct_InvalidID(t *testing.T) {
	h, _ := makeProductHandler(t)

	c, rec := newProductContext(http.MethodDelete, "/products/not-a-uuid", "", map[string]string{"id": "not-a-uuid"})
	err := h.DeleteProduct(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
