package services_test

import (
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repomocks "golang-order-manager-api/internal/repositories/mocks"
	"golang-order-manager-api/internal/services"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeProductService(t *testing.T) (*services.ProductService, *repomocks.MockProductRepository) {
	mockRepo := repomocks.NewMockProductRepository(t)
	svc := services.NewProductService(mockRepo)
	return svc, mockRepo
}

func TestProductGetByID_Success(t *testing.T) {
	svc, mockRepo := makeProductService(t)

	productID := uuid.New()
	product := models.Product{
		ID:          productID,
		Name:        "Test Product",
		Description: "Test description",
		Price:       10.0,
		Stock:       5,
	}

	mockRepo.EXPECT().GetByID(productID).Return(product, nil)

	result, err := svc.GetByID(productID)

	require.NoError(t, err)
	assert.Equal(t, productID, result.ID)
	assert.Equal(t, "Test Product", result.Name)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, 10.0, result.Price)
	assert.Equal(t, 5, result.Stock)
}

func TestProductGetByID_NotFound(t *testing.T) {
	svc, mockRepo := makeProductService(t)

	productID := uuid.New()

	mockRepo.EXPECT().GetByID(productID).Return(models.Product{}, error_codes.ErrProductNotFound)

	_, err := svc.GetByID(productID)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
}

func TestProductCreate_Success(t *testing.T) {
	svc, mockRepo := makeProductService(t)
	product := models.Product{
		Name:        "New Product",
		Description: "New description",
		Price:       20.0,
		Stock:       10,
	}

	mockRepo.EXPECT().Create(product).Return(product, nil)

	result, err := svc.Create(product)
	require.NoError(t, err)
	assert.Equal(t, product.Name, result.Name)
	assert.Equal(t, product.Description, result.Description)
	assert.Equal(t, product.Price, result.Price)
	assert.Equal(t, product.Stock, result.Stock)
}

func TestProductCreate_InvalidPrice(t *testing.T) {
	svc, _ := makeProductService(t)
	product := models.Product{
		Name:        "Invalid Price Product",
		Description: "Description",
		Price:       -5.0,
		Stock:       10,
	}

	result, err := svc.Create(product)
	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrInvalidPrice)
	assert.Equal(t, models.Product{}, result)
}

func TestProductCreate_InvalidStock(t *testing.T) {
	svc, _ := makeProductService(t)
	product := models.Product{
		Name:        "Invalid Stock Product",
		Description: "Description",
		Price:       5.0,
		Stock:       -10,
	}

	result, err := svc.Create(product)
	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrInvalidStock)
	assert.Equal(t, models.Product{}, result)
}

func TestProductUpdate_Success(t *testing.T) {
	svc, mockRepo := makeProductService(t)
	productID := uuid.New()
	product := models.Product{
		ID:          productID,
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       15.0,
		Stock:       8,
	}

	mockRepo.EXPECT().Update(productID, product.Name, product.Description, &product.Price, &product.Stock).Return(product, nil)

	result, err := svc.Update(productID, product.Name, product.Description, &product.Price, &product.Stock)
	require.NoError(t, err)
	assert.Equal(t, productID, result.ID)
	assert.Equal(t, "Updated Product", result.Name)
	assert.Equal(t, "Updated description", result.Description)
	assert.Equal(t, 15.0, result.Price)
	assert.Equal(t, 8, result.Stock)
}

func TestProductUpdate_NotFound(t *testing.T) {
	svc, mockRepo := makeProductService(t)
	productID := uuid.New()
	product := models.Product{
		ID:          productID,
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       15.0,
		Stock:       8,
	}

	mockRepo.EXPECT().Update(productID, product.Name, product.Description, &product.Price, &product.Stock).Return(models.Product{}, error_codes.ErrProductNotFound)

	result, err := svc.Update(productID, product.Name, product.Description, &product.Price, &product.Stock)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
	assert.Equal(t, models.Product{}, result)
}

func TestProductUpdate_InvalidPrice(t *testing.T) {
	svc, _ := makeProductService(t)
	productID := uuid.New()
	product := models.Product{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       -15.0,
		Stock:       8,
	}

	_, err := svc.Update(productID, product.Name, product.Description, &product.Price, &product.Stock)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrInvalidPrice)
}

func TestProductUpdate_InvalidStock(t *testing.T) {
	svc, _ := makeProductService(t)
	productID := uuid.New()
	product := models.Product{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       15.0,
		Stock:       -8,
	}

	_, err := svc.Update(productID, product.Name, product.Description, &product.Price, &product.Stock)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrInvalidStock)
}

func TestProductDelete_Success(t *testing.T) {
	svc, mockRepo := makeProductService(t)
	productID := uuid.New()

	mockRepo.EXPECT().Delete(productID).Return(nil)

	err := svc.Delete(productID)

	require.NoError(t, err)
}

func TestProductDelete_NotFound(t *testing.T) {
	svc, mockRepo := makeProductService(t)
	productID := uuid.New()

	mockRepo.EXPECT().Delete(productID).Return(error_codes.ErrProductNotFound)

	err := svc.Delete(productID)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
}
