package repository

import (
	"database/sql"
	"fmt"
	"golang-order-manager-api/internal/models"
	"golang-order-manager-api/internal/services"

	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepo {
	return UserRepo{db: db}
}

func (repo *UserRepo) CreateUser(username, email, password string) (user models.User, err error) {
	query := `
		INSERT INTO public."users"
		(username,email,password)
		VALUES
		($1,$2,$3)
		RETURNING id;
	`

	hashPassword, err := services.HashPassword(password)
	if err != nil {
		return models.User{}, fmt.Errorf("error to hash password: %v", err)
	}

	row := repo.db.QueryRow(query, username, email, hashPassword)

	err = row.Scan(&user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("error to create user: %v", err)
	}

	user.Email = email
	user.Username = username

	return user, nil
}

func (repo *UserRepo) DeleteUser(id uuid.UUID) (err error) {
	query := `
		UPDATE users SET deleted_at = NOW() WHERE id = $1
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
	//! BROKEN, TODO: Get ID by token, not by body

	if user.Email != "" {
		isEmailAlreadyInUse, err := repo.IsEmailAlreadyInUse(user.Email)
		if err != nil {
			return fmt.Errorf("error to check if email is already in use: %v", err)
		}

		if isEmailAlreadyInUse {
			return fmt.Errorf("email is already in use: %v", user.Email)
		}
	}

	if user.Username != "" {
		isUsernameAlreadyInUse, err := repo.IsUsernameAlreadyInUse(user.Username)
		if err != nil {
			return fmt.Errorf("error to check if username is already in use: %v", err)
		}

		if isUsernameAlreadyInUse {
			return fmt.Errorf("username is already in use: %v", user.Username)
		}
	}

	query := `
		UPDATE users 
		SET 
			username = COALESCE(NULLIF($1, ''), username),
			email = COALESCE(NULLIF($2, ''), email),
			password = COALESCE(NULLIF($3, ''), password)
		WHERE id = $4
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

func (repo *UserRepo) IsEmailAlreadyInUse(email string) (exists bool, err error) {
	query := `
		select exists
		(
		select 1
			from "users"
			where email = $1
		)
	`

	row := repo.db.QueryRow(query, email)

	err = row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error to scan row: %v", err)
	}

	return exists, nil
}

func (repo *UserRepo) IsUsernameAlreadyInUse(username string) (exists bool, err error) {
	query := `
		select exists
		(
			select 1
			from "users"
			where username = $1
		)
	`

	row := repo.db.QueryRow(query, username)

	err = row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error to scan row: %v", err)
	}

	return exists, nil
}
