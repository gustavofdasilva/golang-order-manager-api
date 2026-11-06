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

func TestOrderRepo_GetAll_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	orderA := testhelpers.CreateOrderFixture(t, db, nil)
	_ = testhelpers.CreateOrderFixture(t, db, &orderA.UserID) //Create order to the same user from OrderA
	_ = testhelpers.CreateOrderFixture(t, db, &orderA.UserID) //=

	repo := repository.NewOrderRepo(db)

	result, count, err := repo.GetAllByUserID(orderA.UserID, 10, 0, models.OrderFilter{})

	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(result))
}

func TestOrderRepo_GetByID_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	order := testhelpers.CreateOrderFixture(t, db, nil)

	repo := repository.NewOrderRepo(db)

	result, err := repo.GetByID(order.ID)

	require.NoError(t, err)
	assert.Equal(t, order.ID, result.ID)
	assert.Equal(t, order.Status, result.Status)
	assert.Equal(t, order.UserID, result.UserID)
	assert.Equal(t, order.TotalAmount, result.TotalAmount)
}

func TestOrderRepo_GetByID_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewOrderRepo(db)

	_, err := repo.GetByID(uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrOrderNotFound)
}

func TestOrderRepo_Create_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)
	productA := testhelpers.CreateProductFixture(t, db, 10, 10)
	productB := testhelpers.CreateProductFixture(t, db, 10, 10)

	orderItemA := models.OrderItem{
		ProductID: productA.ID,
		Quantity:  5,
		UnitPrice: 10,
		Subtotal:  50,
	}

	orderItemB := models.OrderItem{
		ProductID: productB.ID,
		Quantity:  5,
		UnitPrice: 10,
		Subtotal:  50,
	}

	order := models.Order{
		UserID:      user.ID,
		Status:      models.OrderStatusPending,
		TotalAmount: 100.0,
		Items: []models.OrderItem{
			orderItemA, orderItemB,
		},
	}

	repo := repository.NewOrderRepo(db)

	result, err := repo.Create(order)
	items, err := repo.CreateItems(result.ID, order.Items)

	require.NoError(t, err)
	assert.Equal(t, order.Status, result.Status)
	assert.Equal(t, order.UserID, result.UserID)
	assert.Equal(t, order.TotalAmount, result.TotalAmount)
	assert.Equal(t, len(order.Items), len(items))
	assert.Equal(t, 2, len(items))

}

func TestOrderRepo_Update_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	order := testhelpers.CreateOrderFixture(t, db, nil)

	repo := repository.NewOrderRepo(db)

	err := repo.UpdateStatus(order.ID, models.OrderStatusPaid)
	require.NoError(t, err)

	updatedOrder, err := repo.GetByID(order.ID)
	require.NoError(t, err)

	assert.Equal(t, models.OrderStatusPaid, updatedOrder.Status)
}

func TestOrderRepo_Update_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewOrderRepo(db)

	err := repo.UpdateStatus(uuid.New(), models.OrderStatusPaid)

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrOrderNotFound)
}
