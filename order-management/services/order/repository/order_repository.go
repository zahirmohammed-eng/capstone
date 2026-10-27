package repository

import (
	"context"
	"errors"

	"order-management/middleware"
	"order-management/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("order not found")

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	List(ctx context.Context) ([]model.Order, error)
	UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) error
}

type gormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) OrderRepository {
	return &gormOrderRepository{db: db}
}

func (r *gormOrderRepository) Create(ctx context.Context, order *model.Order) error {
	err := r.db.WithContext(ctx).Create(order).Error
	if err != nil {
		middleware.GlobalMetrics.IncDBErrors()
		return err
	}
	return nil
}

func (r *gormOrderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		middleware.GlobalMetrics.IncDBErrors()
		return nil, err
	}
	return &o, nil
}

func (r *gormOrderRepository) List(ctx context.Context) ([]model.Order, error) {
	var list []model.Order
	err := r.db.WithContext(ctx).Order("id DESC").Find(&list).Error
	if err != nil {
		middleware.GlobalMetrics.IncDBErrors()
		return nil, err
	}
	return list, nil
}

func (r *gormOrderRepository) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) error {
	res := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		middleware.GlobalMetrics.IncDBErrors()
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
