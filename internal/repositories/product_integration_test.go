package repository_test

import (
	"testing"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepo_GetAll_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	_ = testhelpers.CreateProductFixture(t, db, 100.0, 10)
	_ = testhelpers.CreateProductFixture(t, db, 110.0, 30)
	_ = testhelpers.CreateProductFixture(t, db, 120.0, 20)

	repo := repository.NewProductRepo(db)

	result, count, err := repo.GetAll(10, 0, repository.ProductFilter{})

	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(result))
}

func TestProductRepo_GetByID_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := testhelpers.CreateProductFixture(t, db, 100.0, 10)
	repo := repository.NewProductRepo(db)

	result, err := repo.GetByID(product.ID)

	require.NoError(t, err)
	assert.Equal(t, product.ID, result.ID)
	assert.Equal(t, product.Price, result.Price)
	assert.Equal(t, product.Stock, result.Stock)
}

func TestProductRepo_GetByID_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewProductRepo(db)

	_, err := repo.GetByID(uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
}

func TestProductRepo_Create_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := models.Product{
		Name:        "Test product",
		Description: "Test description",
		Price:       100.0,
		Stock:       10,
	}

	repo := repository.NewProductRepo(db)

	result, err := repo.Create(product)

	require.NoError(t, err)
	assert.Equal(t, product.Name, result.Name)
	assert.Equal(t, product.Description, result.Description)
	assert.Equal(t, product.Price, result.Price)
	assert.Equal(t, product.Stock, result.Stock)
}

func TestProductRepo_Update_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	oldProduct := testhelpers.CreateProductFixture(t, db, 110.0, 20)

	product := models.Product{
		Name:        "Updated test product",
		Description: "Updated test description",
		Price:       100.0,
		Stock:       10,
	}

	repo := repository.NewProductRepo(db)

	result, err := repo.Update(oldProduct.ID, product.Name, product.Description, &product.Price, &product.Stock)

	require.NoError(t, err)
	assert.Equal(t, product.Name, result.Name)
	assert.Equal(t, product.Description, result.Description)
	assert.Equal(t, product.Price, result.Price)
	assert.Equal(t, product.Stock, result.Stock)
}

func TestProductRepo_Update_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := models.Product{
		ID:          uuid.New(),
		Name:        "Updated test product",
		Description: "Updated test description",
		Price:       100.0,
		Stock:       10,
	}

	repo := repository.NewProductRepo(db)

	_, err := repo.Update(product.ID, product.Name, product.Description, &product.Price, &product.Stock)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
}

func TestProductRepo_Delete_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := testhelpers.CreateProductFixture(t, db, 100.0, 10)

	repo := repository.NewProductRepo(db)

	err := repo.Delete(product.ID)

	require.NoError(t, err)
}

func TestProductRepo_Delete_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewProductRepo(db)

	err := repo.Delete(uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrProductNotFound)
}

func TestProductRepo_IncrementStock_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := testhelpers.CreateProductFixture(t, db, 100.0, 10)
	repo := repository.NewProductRepo(db)

	err := repo.IncrementStock(product.ID, 10)
	require.NoError(t, err)

	updatedProduct, err := repo.GetByID(product.ID)
	require.NoError(t, err)

	assert.Equal(t, 20, updatedProduct.Stock)
	assert.Equal(t, product.ID, updatedProduct.ID)
	assert.Equal(t, product.Name, updatedProduct.Name)
	assert.Equal(t, product.Description, updatedProduct.Description)
}

func TestProductRepo_DecrementStock_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := testhelpers.CreateProductFixture(t, db, 100.0, 10)
	repo := repository.NewProductRepo(db)

	err := repo.DecrementStock(product.ID, 5)
	require.NoError(t, err)

	updatedProduct, err := repo.GetByID(product.ID)
	require.NoError(t, err)

	assert.Equal(t, 5, updatedProduct.Stock)
	assert.Equal(t, product.ID, updatedProduct.ID)
	assert.Equal(t, product.Name, updatedProduct.Name)
	assert.Equal(t, product.Description, updatedProduct.Description)
}

func TestProductRepo_DecrementStock_InsufficientStock(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	product := testhelpers.CreateProductFixture(t, db, 100.0, 10)
	repo := repository.NewProductRepo(db)

	err := repo.DecrementStock(product.ID, 20)
	require.Error(t, err)

	assert.ErrorIs(t, err, error_codes.ErrInsufficientStock)
}
