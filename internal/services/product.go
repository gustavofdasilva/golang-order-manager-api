package services

import (
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"

	"github.com/google/uuid"
)

type ProductService interface {
	GetAll(page, limit int, filter repository.ProductFilter) ([]models.Product, int, error)
	GetByID(id uuid.UUID) (models.Product, error)
	Create(p models.Product) (models.Product, error)
	Update(id uuid.UUID, name, description string, price *float64, stock *int) (models.Product, error)
	Delete(id uuid.UUID) error
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{productRepo: productRepo}
}

func (s *productService) GetAll(page, limit int, filter repository.ProductFilter) ([]models.Product, int, error) {
	offset := (page - 1) * limit
	return s.productRepo.GetAll(limit, offset, filter)
}

func (s *productService) GetByID(id uuid.UUID) (models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *productService) Create(p models.Product) (models.Product, error) {
	if p.Price < 0 {
		return models.Product{}, error_codes.ErrInvalidPrice
	}
	if p.Stock < 0 {
		return models.Product{}, error_codes.ErrInvalidStock
	}
	return s.productRepo.Create(p)
}

func (s *productService) Update(id uuid.UUID, name, description string, price *float64, stock *int) (models.Product, error) {
	if price != nil && *price < 0 {
		return models.Product{}, error_codes.ErrInvalidPrice
	}
	if stock != nil && *stock < 0 {
		return models.Product{}, error_codes.ErrInvalidStock
	}
	return s.productRepo.Update(id, name, description, price, stock)
}

func (s *productService) Delete(id uuid.UUID) error {
	return s.productRepo.Delete(id)
}
