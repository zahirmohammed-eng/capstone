package service

import (
	"context"
	"errors"
	"fmt"

	"order-management/model"
	"order-management/services/order/client"
	"order-management/services/order/repository"
)

var ErrInvalidStatus = errors.New("invalid status: must be CREATED, CONFIRMED, or CANCELLED")

type OrderService interface {
	CreateOrder(ctx context.Context, req model.CreateOrderRequest) (*model.Order, error)
	GetOrder(ctx context.Context, id int64) (*model.Order, error)
	ListOrders(ctx context.Context) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status model.OrderStatus) error
}

type orderService struct {
	repo          repository.OrderRepository
	productClient client.ProductClient
}

func NewOrderService(repo repository.OrderRepository, pc client.ProductClient) OrderService {
	return &orderService{
		repo:          repo,
		productClient: pc,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (*model.Order, error) {
	// Call resilient client to fetch live product price from Product Service
	unitPrice, err := s.productClient.GetProductPrice(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product price: %w", err)
	}

	order := &model.Order{
		CustomerID:  req.CustomerID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		TotalAmount: float64(req.Quantity) * unitPrice,
		Status:      model.StatusCreated,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context) ([]model.Order, error) {
	return s.repo.List(ctx)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id int64, status model.OrderStatus) error {
	switch status {
	case model.StatusCreated, model.StatusConfirmed, model.StatusCancelled:
		return s.repo.UpdateStatus(ctx, id, status)
	default:
		return ErrInvalidStatus
	}
}
