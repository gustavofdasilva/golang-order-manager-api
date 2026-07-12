package repository

import (
	"database/sql"
	"fmt"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"

	"github.com/google/uuid"
)

type ProductFilter struct {
	Name     *string
	MinPrice *float64
	MaxPrice *float64
	InStock  *bool
}

type ProductRepository interface {
	GetAll(limit, offset int, f ProductFilter) ([]models.Product, int, error)
	GetByID(id uuid.UUID) (models.Product, error)
	Create(p models.Product) (models.Product, error)
	Update(id uuid.UUID, name, description string, price *float64, stock *int) (models.Product, error)
	Delete(id uuid.UUID) error
	DecrementStock(productID uuid.UUID, quantity int) error
	IncrementStock(productID uuid.UUID, quantity int) error
}

type ProductRepo struct {
	db DBTX
}

func NewProductRepo(db DBTX) ProductRepository {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) GetAll(limit, offset int, f ProductFilter) ([]models.Product, int, error) {
	filterClause := `
		AND ($1::text    IS NULL OR name  ILIKE '%' || $1 || '%')
		AND ($2::numeric IS NULL OR price >= $2)
		AND ($3::numeric IS NULL OR price <= $3)
		AND ($4::boolean IS NULL OR stock > 0)
	`

	var total int
	countQuery := `SELECT COUNT(*) FROM items WHERE deleted_at IS NULL ` + filterClause
	if err := r.db.QueryRow(countQuery, f.Name, f.MinPrice, f.MaxPrice, f.InStock).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting products: %v", err)
	}

	query := `
		SELECT id, name, description, price, stock, created_at, updated_at
		FROM items
		WHERE deleted_at IS NULL ` + filterClause + `
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.db.Query(query, f.Name, f.MinPrice, f.MaxPrice, f.InStock, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying products: %v", err)
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("error scanning product: %v", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating products: %v", err)
	}
	return products, total, nil
}

func (r *ProductRepo) GetByID(id uuid.UUID) (models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, created_at, updated_at
		FROM items
		WHERE id = $1 AND deleted_at IS NULL
	`
	var p models.Product
	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Product{}, error_codes.ErrProductNotFound
		}
		return models.Product{}, fmt.Errorf("error scanning product: %v", err)
	}
	return p, nil
}

func (r *ProductRepo) Create(p models.Product) (models.Product, error) {
	query := `
		INSERT INTO items (name, description, price, stock)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(query, p.Name, p.Description, p.Price, p.Stock).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.Product{}, fmt.Errorf("error creating product: %v", err)
	}
	return p, nil
}

// ! Update should receive models.Product instead of individual fields, but for now we keep it as is to match the service layer.
func (r *ProductRepo) Update(id uuid.UUID, name, description string, price *float64, stock *int) (models.Product, error) {
	query := `
		UPDATE items
		SET
			name        = COALESCE(NULLIF($1, ''), name),
			description = COALESCE(NULLIF($2, ''), description),
			price       = COALESCE($3, price),
			stock       = COALESCE($4, stock),
			updated_at  = NOW()
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, name, description, price, stock, created_at, updated_at
	`
	var p models.Product
	err := r.db.QueryRow(query, name, description, price, stock, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Product{}, error_codes.ErrProductNotFound
		}
		return models.Product{}, fmt.Errorf("error updating product: %v", err)
	}
	return p, nil
}

func (r *ProductRepo) Delete(id uuid.UUID) error {
	query := `UPDATE items SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if affected == 0 {
		return error_codes.ErrProductNotFound
	}
	return nil
}

func (r *ProductRepo) DecrementStock(productID uuid.UUID, quantity int) error {
	// Atomic check-and-decrement: fails if stock < quantity
	query := `
		UPDATE items SET stock = stock - $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND stock >= $1
	`
	result, err := r.db.Exec(query, quantity, productID)
	if err != nil {
		return fmt.Errorf("error decrementing stock: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if affected == 0 {
		return error_codes.ErrInsufficientStock
	}
	return nil
}

func (r *ProductRepo) IncrementStock(productID uuid.UUID, quantity int) error {
	query := `UPDATE items SET stock = stock + $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(query, quantity, productID)
	if err != nil {
		return fmt.Errorf("error incrementing stock: %v", err)
	}
	return nil
}
