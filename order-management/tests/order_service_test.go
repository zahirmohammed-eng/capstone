package tests

import (
	"context"
	"errors"
	"testing"

	"order-management/model"
	"order-management/services/order/repository"
	"order-management/services/order/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockProductClient struct {
	price float64
	err   error
}

func (m *mockProductClient) GetProductPrice(ctx context.Context, productID int64) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.price, nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	_ = db.AutoMigrate(&model.Order{})
	return db
}

func TestCreateOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewGormOrderRepository(db)
	pc := &mockProductClient{price: 49.99}
	svc := service.NewOrderService(repo, pc)

	req := model.CreateOrderRequest{
		CustomerID: 101,
		ProductID:  501,
		Quantity:   2,
	}

	order, err := svc.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if order.ID == 0 {
		t.Errorf("expected positive order ID, got %d", order.ID)
	}
	if order.UnitPrice != 49.99 {
		t.Errorf("expected unit price 49.99, got %f", order.UnitPrice)
	}
	if order.TotalAmount != 99.98 {
		t.Errorf("expected total amount 99.98, got %f", order.TotalAmount)
	}
}

func TestUpdateOrderStatus_TableDriven(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewGormOrderRepository(db)
	pc := &mockProductClient{price: 10.0}
	svc := service.NewOrderService(repo, pc)

	order, _ := svc.CreateOrder(context.Background(), model.CreateOrderRequest{
		CustomerID: 101,
		ProductID:  501,
		Quantity:   1,
	})

	tests := []struct {
		name        string
		orderID     int64
		status      model.OrderStatus
		expectedErr error
	}{
		{
			name:        "Valid transition to CONFIRMED",
			orderID:     order.ID,
			status:      model.StatusConfirmed,
			expectedErr: nil,
		},
		{
			name:        "Valid transition to CANCELLED",
			orderID:     order.ID,
			status:      model.StatusCancelled,
			expectedErr: nil,
		},
		{
			name:        "Invalid status transition",
			orderID:     order.ID,
			status:      model.OrderStatus("UNKNOWN"),
			expectedErr: service.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateOrderStatus(context.Background(), tt.orderID, tt.status)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
