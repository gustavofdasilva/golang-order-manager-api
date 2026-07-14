package services_test

import (
	"fmt"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repomocks "golang-order-manager-api/internal/repositories/mocks"
	"golang-order-manager-api/internal/security"
	"golang-order-manager-api/internal/services"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testJWTSecret            = "test-secret-key"
	testRefreshTokenDuration = 7 * 24 * time.Hour
)

type authServiceMocks struct {
	userRepo *repomocks.MockUserRepository
	authRepo *repomocks.MockAuthRepository
}

func makeAuthService(t *testing.T) (services.AuthService, authServiceMocks) {
	m := authServiceMocks{
		userRepo: repomocks.NewMockUserRepository(t),
		authRepo: repomocks.NewMockAuthRepository(t),
	}
	svc := services.NewAuthService(m.userRepo, m.authRepo, testJWTSecret, testRefreshTokenDuration)
	return svc, m
}

func TestRegister_Success(t *testing.T) {
	svc, m := makeAuthService(t)

	name := "john"
	email := "john@example.com"
	password := "plaintext123"

	m.userRepo.EXPECT().IsEmailAlreadyInUse(email, mock.Anything).Return(false, nil)
	m.userRepo.EXPECT().IsUsernameAlreadyInUse(name, mock.Anything).Return(false, nil)
	m.userRepo.EXPECT().
		CreateUser(mock.MatchedBy(func(u models.User) bool {
			return u.Email == email &&
				u.Username == name &&
				u.Password != "plaintext123"
		})).
		Return(models.User{ID: uuid.New(), Email: email, Username: name}, nil)

	result, err := svc.Register(name, email, password)

	require.NoError(t, err)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, name, result.Username)
}

func TestAuthRegister_EmailAlreadyInUse(t *testing.T) {
	svc, m := makeAuthService(t)

	email := "john@email.com"
	username := "John"
	password := "plaintext123"

	m.userRepo.EXPECT().IsEmailAlreadyInUse(email, mock.Anything).Return(true, nil)

	_, err := svc.Register(username, email, password)

	assert.ErrorIs(t, err, error_codes.ErrEmailAlreadyInUse)
}

func TestAuthRegister_UsernameAlreadyInUse(t *testing.T) {
	svc, m := makeAuthService(t)

	email := "john@email.com"
	username := "John"
	password := "plaintext123"

	m.userRepo.EXPECT().IsEmailAlreadyInUse(email, mock.Anything).Return(false, nil)
	m.userRepo.EXPECT().IsUsernameAlreadyInUse(username, mock.Anything).Return(true, nil)

	_, err := svc.Register(username, email, password)

	assert.ErrorIs(t, err, error_codes.ErrUsernameAlreadyInUse)
}

func TestLogin_Success(t *testing.T) {
	svc, m := makeAuthService(t)

	hashedPassword, _ := security.HashPassword("correct-password")
	user := models.User{
		ID:       uuid.New(),
		Email:    "john@example.com",
		Password: hashedPassword,
	}

	m.userRepo.EXPECT().GetByEmail("john@example.com").Return(user, nil)
	m.authRepo.EXPECT().
		SaveRefreshToken(user.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(nil)

	token, refreshToken, result, err := svc.Login("john@example.com", "correct-password")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, refreshToken)
	assert.Empty(t, result.Password) // senha limpa antes de retornar
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, m := makeAuthService(t)

	hashedPassword, _ := security.HashPassword("correct-password")
	user := models.User{
		ID:       uuid.New(),
		Email:    "john@example.com",
		Password: hashedPassword,
	}

	m.userRepo.EXPECT().GetByEmail("john@example.com").Return(user, nil)

	token, refreshToken, result, err := svc.Login("john@example.com", "wrong-password")

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrInvalidCredentials)
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	assert.Empty(t, result.Password) // senha limpa antes de retornar
}

func TestLogin_NotFound(t *testing.T) {
	svc, m := makeAuthService(t)

	m.userRepo.EXPECT().GetByEmail("john@example.com").Return(models.User{}, error_codes.ErrUserNotFound)

	token, refreshToken, result, err := svc.Login("john@example.com", "password")

	require.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrUserNotFound)
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	assert.Empty(t, result.Password) // senha limpa antes de retornar
}

func TestRefresh_Success(t *testing.T) {
	svc, m := makeAuthService(t)

	userID := uuid.New()
	oldRefreshToken := "valid-refresh-token"
	hashedOldToken := security.HashToken(oldRefreshToken)

	user := models.User{ID: userID, Email: "john@example.com"}

	// GetRefreshToken recebe o hash, não o token cru
	m.authRepo.EXPECT().
		GetRefreshToken(hashedOldToken).
		Return(userID, time.Time{}, time.Now().Add(time.Hour), nil)

	m.userRepo.EXPECT().GetByID(userID).Return(user, nil)

	m.authRepo.EXPECT().RevokeRefreshToken(hashedOldToken).Return(nil)

	m.authRepo.EXPECT().
		SaveRefreshToken(userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(nil)

	token, newRefreshToken, result, err := svc.Refresh(oldRefreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken) // token rotacionado
	assert.Equal(t, userID, result.ID)
}

func TestRefresh_TokenExpired(t *testing.T) {
	svc, m := makeAuthService(t)

	refreshToken := "expired-token"
	hashedToken := security.HashToken(refreshToken)

	m.authRepo.EXPECT().
		GetRefreshToken(hashedToken).
		Return(uuid.Nil, time.Time{}, time.Now().Add(-time.Hour), nil) // expirado

	_, _, _, err := svc.Refresh(refreshToken)

	assert.ErrorIs(t, err, error_codes.ErrRefreshTokenExpired)
}

func TestRefresh_TokenRevoked(t *testing.T) {
	svc, m := makeAuthService(t)

	refreshToken := "revoked-token"
	hashedToken := security.HashToken(refreshToken)

	m.authRepo.EXPECT().
		GetRefreshToken(hashedToken).
		Return(uuid.Nil, time.Now(), time.Now().Add(time.Hour), nil) // revokedAt preenchido

	_, _, _, err := svc.Refresh(refreshToken)

	assert.ErrorIs(t, err, error_codes.ErrRefreshTokenExpired)
}

func TestLogout_Success(t *testing.T) {
	svc, m := makeAuthService(t)

	refreshToken := "valid-token"
	hashedToken := security.HashToken(refreshToken)

	m.authRepo.EXPECT().RevokeRefreshToken(hashedToken).Return(nil)

	err := svc.Logout(refreshToken)

	require.NoError(t, err)
}

func TestLogout_RepoError(t *testing.T) {
	svc, m := makeAuthService(t)

	refreshToken := "valid-token"
	hashedToken := security.HashToken(refreshToken)

	m.authRepo.EXPECT().
		RevokeRefreshToken(hashedToken).
		Return(fmt.Errorf("db error"))

	err := svc.Logout(refreshToken)

	assert.Error(t, err)
}

func TestLogoutAll_Success(t *testing.T) {
	svc, m := makeAuthService(t)

	userID := uuid.New()
	m.authRepo.EXPECT().RevokeAllRefreshTokens(userID).Return(nil)

	err := svc.LogoutAll(userID)

	require.NoError(t, err)
}
