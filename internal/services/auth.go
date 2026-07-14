package services

import (
	"fmt"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/security"
	"time"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(username string, email string, password string) (models.User, error)
	Login(email string, password string) (string, string, models.User, error)
	Refresh(refreshToken string) (string, string, models.User, error)
	Logout(refreshToken string) error
	LogoutAll(userID uuid.UUID) error
}

type authService struct {
	userRepo             repository.UserRepository
	authRepo             repository.AuthRepository
	refreshTokenDuration time.Duration
	jwtSecret            string
}

func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository, jwtSecret string, refreshTokenDuration time.Duration) AuthService {
	return &authService{
		userRepo:             userRepo,
		authRepo:             authRepo,
		refreshTokenDuration: refreshTokenDuration,
		jwtSecret:            jwtSecret,
	}
}

// ? Should CreateUser be part of AuthService or UserService? I think it should be part of UserService, but for now, I'll leave it here.
func (s *authService) Register(username string, email string, password string) (models.User, error) {
	exists, err := s.userRepo.IsEmailAlreadyInUse(email, nil)
	if err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, error_codes.ErrEmailAlreadyInUse
	}

	exists, err = s.userRepo.IsUsernameAlreadyInUse(username, nil)
	if err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, error_codes.ErrUsernameAlreadyInUse
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	newUser := models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}

	user, err := s.userRepo.CreateUser(newUser)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (s *authService) Login(email string, password string) (string, string, models.User, error) {

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", "", models.User{}, err
	}

	err = security.ComparePassword(user.Password, password)
	if err != nil {
		return "", "", models.User{}, error_codes.ErrInvalidCredentials
	}

	token, err := security.GenerateToken(user, s.jwtSecret)
	if err != nil {
		return "", "", models.User{}, err
	}

	refreshToken, err := security.GenerateRefreshToken()
	if err != nil {
		return "", "", models.User{}, err
	}

	err = s.SaveRefreshToken(user.ID, refreshToken)
	if err != nil {
		return "", "", models.User{}, err
	}

	user.Password = ""

	return token, refreshToken, user, nil
}

func (s *authService) Refresh(refreshToken string) (string, string, models.User, error) {

	userID, err := s.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		return "", "", models.User{}, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", "", models.User{}, err
	}

	token, err := security.GenerateToken(user, s.jwtSecret)
	if err != nil {
		return "", "", models.User{}, err
	}

	hashedToken := security.HashToken(refreshToken)
	err = s.authRepo.RevokeRefreshToken(hashedToken)
	if err != nil {
		return "", "", models.User{}, err
	}

	newRefreshToken, err := security.GenerateRefreshToken()
	if err != nil {
		return "", "", models.User{}, err
	}

	err = s.SaveRefreshToken(user.ID, newRefreshToken)
	if err != nil {
		return "", "", models.User{}, err
	}

	return token, newRefreshToken, user, nil
}

func (s *authService) Logout(refreshToken string) error {

	hashedToken := security.HashToken(refreshToken)
	err := s.authRepo.RevokeRefreshToken(hashedToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) LogoutAll(userID uuid.UUID) error {
	err := s.authRepo.RevokeAllRefreshTokens(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) SaveRefreshToken(userID uuid.UUID, refreshToken string) error {

	expiresAt := time.Now().Add(s.refreshTokenDuration)

	hashedToken := security.HashToken(refreshToken)

	err := s.authRepo.SaveRefreshToken(userID, hashedToken, expiresAt)
	if err != nil {
		return fmt.Errorf("error saving refresh token: %v", err)
	}

	return nil
}

func (s *authService) GetUserIDByRefreshToken(refreshToken string) (uuid.UUID, error) {

	hashedToken := security.HashToken(refreshToken)

	userID, revokedAt, expiresAt, err := s.authRepo.GetRefreshToken(hashedToken)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error retrieving refresh token: %v", err)
	}

	if !revokedAt.IsZero() || time.Now().After(expiresAt) {
		return uuid.Nil, error_codes.ErrRefreshTokenExpired
	}

	return userID, nil
}
