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

func (s *ProductService) GetAll() ([]models.Product, error) {
	return s.productRepo.GetAll()
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

	//TODO: Woudnt be better to just update in the repo like 'COALESCE($1, name)' and avoid the extra query to get the current values? This way we also avoid
	current, err := s.productRepo.GetByID(id)
	if err != nil {
		return models.Product{}, err
	}

	if name != "" {
		current.Name = name
	}
	if description != "" {
		current.Description = description
	}
	if price != nil {
		if *price < 0 { //TODO: Free products allowed?
			return models.Product{}, error_codes.ErrInvalidPrice
		}
		current.Price = *price
	}
	if stock != nil {
		if *stock < 0 {
			return models.Product{}, error_codes.ErrInvalidStock
		}
		current.Stock = *stock
	}

	return s.productRepo.Update(current)
}

func (s *ProductService) Delete(id uuid.UUID) error {
	return s.productRepo.Delete(id)
}
