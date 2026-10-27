package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"order-management/model"
	"order-management/services/order/repository"
	"order-management/services/order/service"
)

func Test100ConcurrentOrderRequests(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewGormOrderRepository(db)
	pc := &mockProductClient{price: 20.0}
	svc := service.NewOrderService(repo, pc)

	const totalRequests = 100
	var wg sync.WaitGroup
	var successCount uint64

	wg.Add(totalRequests)
	for i := 0; i < totalRequests; i++ {
		go func(customerID int) {
			defer wg.Done()
			req := model.CreateOrderRequest{
				CustomerID: int64(customerID),
				ProductID:  501,
				Quantity:   1,
			}
			order, err := svc.CreateOrder(context.Background(), req)
			if err == nil && order.ID > 0 {
				atomic.AddUint64(&successCount, 1)
			}
		}(1000 + i)
	}

	wg.Wait()
	if successCount != totalRequests {
		t.Fatalf("expected %d orders created, got %d", totalRequests, successCount)
	}
}
