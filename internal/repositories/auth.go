package repository

import (
	"database/sql"
	"fmt"
	"golang-order-manager-api/internal/models"
	"golang-order-manager-api/internal/services"
)

type AuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) AuthRepo {
	return AuthRepo{db: db}
}

func (repo *AuthRepo) RegisterUser(username, email, password string) (user models.User, err error) {
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

func (repo *AuthRepo) LoginUser(email, password string) (models.User, error) {
	query := `
		SELECT id, username, email, password FROM "users" u WHERE u.email = $1
	`

	row := repo.db.QueryRow(query, email)

	user := models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("error to scan row: %v", err)
	}

	err = services.ComparePassword(user.Password, password)
	if err != nil {
		return models.User{}, fmt.Errorf("error to compare passwords: %v", err)
	}

	return user, nil
}
