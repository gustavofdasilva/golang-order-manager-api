package services

import (
	"errors"
	"fmt"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/security"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`

	jwt.RegisteredClaims
}

func (s *AuthService) Register(username string, email string, password string) (models.User, error) {

	exists, err := s.userRepo.IsEmailAlreadyInUse(email, nil)
	if err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, errors.New("email already in use")
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

func (s *AuthService) Login(email string, password string) (string, models.User, error) {

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", models.User{}, err
	}

	err = security.ComparePassword(user.Password, password)
	if err != nil {
		return "", models.User{}, errors.New("invalid credentials")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return "", models.User{}, err
	}

	user.Password = ""

	return token, user, nil
}

func (s *AuthService) GenerateToken(user models.User) (string, error) {

	claims := Claims{
		UserID: user.ID.String(),
		Email:  user.Email,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(24 * time.Hour),
			),

			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte(s.jwtSecret),
	)

	if err != nil {
		return "", fmt.Errorf(
			"error generating token: %v",
			err,
		)
	}

	return tokenString, nil
}

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwtSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
