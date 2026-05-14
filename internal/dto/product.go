package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductRequest struct {
	Name        string  `json:"name" example:"Running Shoes"`
	Description string  `json:"description,omitempty" example:"High performance running shoes"`
	Price       float64 `json:"price" example:"129.99"`
	Stock       int     `json:"stock" example:"50"`
}

type UpdateProductRequest struct {
	Name        string   `json:"name,omitempty" example:"Running Shoes"`
	Description string   `json:"description,omitempty" example:"High performance running shoes"`
	Price       *float64 `json:"price,omitempty" example:"129.99"`
	Stock       *int     `json:"stock,omitempty" example:"50"`
}

type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
}
