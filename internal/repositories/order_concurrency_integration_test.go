package repository_test

import (
	"errors"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	repomocks "golang-order-manager-api/internal/repositories/mocks"
	"golang-order-manager-api/internal/services"
	"golang-order-manager-api/internal/testhelpers"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateStatus_ConcurrentPay(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	const concurrentUpdates = 10

	order, products := testhelpers.CreateOrderFixtureManyStock(t, db, nil)

	repo := repository.NewOrderRepo(db)
	productRepo := repository.NewProductRepo(db)
	txFactory := repository.NewTxFactory(db)
	cache := repomocks.NewMockOrderCache(t)
	cache.EXPECT().
		DeleteOrder(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	svc := services.NewOrderService(
		txFactory,
		repo,
		cache,
	)

	var ready, completed sync.WaitGroup
	var success, failed atomic.Int32
	start := make(chan struct{})
	errChan := make(chan error, concurrentUpdates)

	ready.Add(concurrentUpdates)
	completed.Add(concurrentUpdates)

	for i := 0; i < concurrentUpdates; i++ {
		go func() {
			defer completed.Done()
			ready.Done()
			<-start

			_, err := svc.UpdateStatus(order.ID, order.UserID, models.OrderStatusPaid)
			switch {
			case err == nil:
				success.Add(1)
			case errors.Is(err, error_codes.ErrOrderStatusTransition):
				failed.Add(1)
			default:
				errChan <- err
			}
		}()
	}

	ready.Wait()
	close(start)
	completed.Wait()
	close(errChan)

	for _, item := range order.Items {
		preCommitItem := products[item.ProductID]

		itemDB, err := productRepo.GetByID(item.ProductID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if preCommitItem.Stock-item.Quantity != itemDB.Stock {
			t.Error("db stock error")
		}
	}

	for err := range errChan {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, int32(1), success.Load())
	assert.Equal(t, int32(concurrentUpdates-1), failed.Load())
}
