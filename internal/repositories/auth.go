package repository

import (
	"database/sql"
	"fmt"
	error_codes "golang-order-manager-api/internal/errors"
	"time"

	"github.com/google/uuid"
)

type AuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) AuthRepo {
	return AuthRepo{db: db}
}

func (repo *AuthRepo) SaveRefreshToken(userID uuid.UUID, refreshToken string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := repo.db.Exec(query, userID, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("error to save refresh token: %v", err)
	}

	return nil
}

func (repo *AuthRepo) GetRefreshToken(refreshToken string) (uuid.UUID, time.Time, time.Time, error) {
	query := `
		SELECT user_id, token, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var userID uuid.UUID
	var expiresAt time.Time
	var revokedAt sql.NullTime

	err := repo.db.QueryRow(query, refreshToken).Scan(&userID, &refreshToken, &expiresAt, &revokedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, time.Time{}, time.Time{}, error_codes.ErrRefreshTokenNotFound
		}
		return uuid.Nil, time.Time{}, time.Time{}, fmt.Errorf("error to scan row: %v", err)
	}

	return userID, revokedAt.Time, expiresAt, nil
}

func (repo *AuthRepo) RevokeRefreshToken(refreshToken string) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE token = $1
	`

	result, err := repo.db.Exec(query, refreshToken)
	if err != nil {
		return fmt.Errorf("error to revoke refresh token: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return error_codes.ErrRefreshTokenNotFound
	}

	return nil
}

func (repo *AuthRepo) RevokeAllRefreshTokens(userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE user_id = $1
	`

	result, err := repo.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error to revoke all refresh tokens: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return error_codes.ErrRefreshTokenNotFound
	}

	return nil
}
