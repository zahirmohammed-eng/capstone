package service

import (
	"context"

	"order-management/model"
	"order-management/services/product/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error)
	GetProduct(ctx context.Context, id int64) (*model.Product, error)
	ListProducts(ctx context.Context) ([]model.Product, error)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	p := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context) ([]model.Product, error) {
	return s.repo.List(ctx)
}
