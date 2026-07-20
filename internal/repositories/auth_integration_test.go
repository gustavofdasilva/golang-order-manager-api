package repository_test

import (
	"testing"
	"time"

	error_codes "golang-order-manager-api/internal/errors"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/security"
	"golang-order-manager-api/internal/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRepo_SaveRefreshToken_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)

	repo := repository.NewAuthRepo(db)

	token, err := security.GenerateToken(user, "secret")
	require.NoError(t, err)

	err = repo.SaveRefreshToken(user.ID, token, time.Now().Add(time.Hour*24))
	require.NoError(t, err)
}

func TestAuthRepo_GetRefreshToken_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)

	repo := repository.NewAuthRepo(db)

	token, err := security.GenerateRefreshToken()
	require.NoError(t, err)

	err = repo.SaveRefreshToken(user.ID, token, time.Now().Add(time.Hour*24))
	require.NoError(t, err)

	userID, _, _, err := repo.GetRefreshToken(token)
	require.NoError(t, err)

	assert.Equal(t, user.ID, userID)
}

func TestAuthRepo_GetRefreshToken_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewAuthRepo(db)

	token, err := security.GenerateRefreshToken()
	require.NoError(t, err)

	_, _, _, err = repo.GetRefreshToken(token)
	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrRefreshTokenNotFound)
}

func TestAuthRepo_RevokeRefreshToken_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)

	repo := repository.NewAuthRepo(db)

	token, err := security.GenerateRefreshToken()
	require.NoError(t, err)

	err = repo.SaveRefreshToken(user.ID, token, time.Now().Add(time.Hour*24))
	require.NoError(t, err)

	err = repo.RevokeRefreshToken(token)
	require.NoError(t, err)
}

func TestAuthRepo_RevokeRefreshToken_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewAuthRepo(db)

	token, err := security.GenerateRefreshToken()
	require.NoError(t, err)

	err = repo.RevokeRefreshToken(token)
	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrRefreshTokenNotFound)
}

func TestAuthRepo_RevokeAllRefreshToken_Success(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	user := testhelpers.CreateUserFixture(t, db)

	repo := repository.NewAuthRepo(db)

	token, err := security.GenerateRefreshToken()
	require.NoError(t, err)

	err = repo.SaveRefreshToken(user.ID, token, time.Now().Add(time.Hour*24))
	require.NoError(t, err)

	err = repo.RevokeAllRefreshTokens(user.ID)
	require.NoError(t, err)
}

func TestAuthRepo_RevokeAllRefreshToken_NotFound(t *testing.T) {
	db := testDB
	t.Cleanup(func() {
		ClearDatabase(db)
	})

	repo := repository.NewAuthRepo(db)

	err := repo.RevokeAllRefreshTokens(uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrRefreshTokenNotFound)
}
