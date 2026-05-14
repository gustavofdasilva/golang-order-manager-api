package services

import (
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"

	"github.com/google/uuid"
)

type ProductService struct {
	productRepo *repository.ProductRepo
}

func NewProductService(productRepo *repository.ProductRepo) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetAll(page, limit int) ([]models.Product, int, error) {
	offset := (page - 1) * limit
	return s.productRepo.GetAll(limit, offset)
}

func (s *ProductService) GetByID(id uuid.UUID) (models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *ProductService) Create(p models.Product) (models.Product, error) {
	if p.Price < 0 { //TODO: Free products allowed?
		return models.Product{}, error_codes.ErrInvalidPrice
	}
	if p.Stock < 0 {
		return models.Product{}, error_codes.ErrInvalidStock
	}
	return s.productRepo.Create(p)
}

func (s *ProductService) Update(id uuid.UUID, name, description string, price *float64, stock *int) (models.Product, error) {
	if price != nil && *price < 0 {
		return models.Product{}, error_codes.ErrInvalidPrice
	}
	if stock != nil && *stock < 0 {
		return models.Product{}, error_codes.ErrInvalidStock
	}
	return s.productRepo.Update(id, name, description, price, stock)
}

func (s *ProductService) Delete(id uuid.UUID) error {
	return s.productRepo.Delete(id)
}
