package repository

import (
	"database/sql"
	"fmt"
	"golang-order-manager-api/internal/models"

	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepo {
	return UserRepo{db: db}
}

func (repo *UserRepo) GetByEmail(email string) (user models.User, err error) {
	query := `
		SELECT id, username, email, password
		FROM public."users"
		WHERE email = $1
	`

	row := repo.db.QueryRow(query, email)

	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("error to scan row: %v", err)
	}

	return user, nil
}

func (repo *UserRepo) GetByID(id uuid.UUID) (user models.User, err error) {
	query := `
		SELECT id, username, email, password
		FROM public."users"
		WHERE id = $1
		AND deleted_at IS NULL
	`

	row := repo.db.QueryRow(query, id)

	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("error to scan row: %v", err)
	}

	return user, nil
}

func (repo *UserRepo) CreateUser(user models.User) (models.User, error) {
	query := `
		INSERT INTO public."users"
		(username,email,password)
		VALUES
		($1,$2,$3)
		RETURNING id;
	`

	row := repo.db.QueryRow(query, user.Username, user.Email, user.Password)

	err := row.Scan(&user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("error to create user: %v", err)
	}

	return user, nil
}

func (repo *UserRepo) DeleteUserByID(id uuid.UUID) (err error) {
	query := `
		UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := repo.db.Exec(query, id)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found for the given id: %v", id)
	}

	return nil
}

func (repo *UserRepo) UpdateUser(user models.User) (err error) {
	query := `
		UPDATE users 
		SET 
			username = COALESCE(NULLIF($1, ''), username),
			email = COALESCE(NULLIF($2, ''), email),
			password = COALESCE(NULLIF($3, ''), password)
		WHERE id = $4
		AND deleted_at IS NULL
	`

	result, err := repo.db.Exec(query, user.Username, user.Email, user.Password, user.ID)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found for the given id: %v", user.ID)
	}

	return nil
}

func (repo *UserRepo) IsEmailAlreadyInUse(email string, excludedID *uuid.UUID) (exists bool, err error) {
	query := `
		SELECT EXISTS
		(
			SELECT 1
			FROM "users"
			WHERE email = $1
			AND ($2::uuid IS NULL OR id != $2::uuid)
			AND deleted_at IS NULL
		)
	`

	row := repo.db.QueryRow(query, email, excludedID)

	err = row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error to scan row: %v", err)
	}

	return exists, nil
}

func (repo *UserRepo) IsUsernameAlreadyInUse(username string, excludedID *uuid.UUID) (exists bool, err error) {
	query := `
		SELECT EXISTS
		(
			SELECT 1
			FROM "users"
			WHERE username = $1
			AND ($2::uuid IS NULL OR id != $2::uuid)
			AND deleted_at IS NULL
		)
	`

	row := repo.db.QueryRow(query, username, excludedID)

	err = row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error to scan row: %v", err)
	}

	return exists, nil
}
