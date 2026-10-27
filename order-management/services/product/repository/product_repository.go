package repository

import (
	"context"
	"errors"

	"order-management/model"
	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context) ([]model.Product, error)
	SeedDefaults(ctx context.Context) error
}

type gormProductRepository struct {
	db *gorm.DB
}

func NewGormProductRepository(db *gorm.DB) ProductRepository {
	return &gormProductRepository{db: db}
}

func (r *gormProductRepository) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *gormProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	var p model.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *gormProductRepository) List(ctx context.Context) ([]model.Product, error) {
	var list []model.Product
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *gormProductRepository) SeedDefaults(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&model.Product{}).Count(&count)
	if count == 0 {
		defaults := []model.Product{
			{ID: 501, Name: "Ergonomic Mechanical Keyboard", Description: "Wireless mechanical keyboard", Price: 99.99, Stock: 50},
			{ID: 502, Name: "Ultra HD Monitor 27-inch", Description: "4K 144Hz Gaming Monitor", Price: 299.50, Stock: 25},
			{ID: 503, Name: "Precision Gaming Mouse", Description: "High DPI optical sensor", Price: 49.00, Stock: 100},
		}
		return r.db.WithContext(ctx).Create(&defaults).Error
	}
	return nil
}
