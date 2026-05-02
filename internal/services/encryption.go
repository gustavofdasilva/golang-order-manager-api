package services

import (
	"fmt"
	"time"

	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func CreateJWTToken(user models.User) (*models.TokenResponse, error) {
	secretKey := config.SECRET_KEY
	exp := time.Now().Add(time.Hour * time.Duration(config.TOKEN_EXPIRATION_HOURS)).Unix()

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user": user,
			"iat":  time.Now().Unix(),
			"exp":  exp,
		})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, fmt.Errorf("error signing token: %w", err)
	}

	tokenResponse := models.TokenResponse{
		Token:    tokenString,
		Exp:      exp,
		Iat:      time.Now().Unix(),
		UserInfo: user,
	}

	return &tokenResponse, nil
}

func VerifyToken(accessToken string) (bool, error) {
	secretKey := config.SECRET_KEY

	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return false, err
	}

	return token.Valid, nil
}

func ReadTokenToUser(tokenString string) (models.User, error) {
	secretKey := config.SECRET_KEY

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return models.User{}, err
	}

	userData, ok := claims["user"].(map[string]interface{})
	if !ok {
		return models.User{}, fmt.Errorf("invalid token claims")
	}

	idStr, ok := userData["id"].(string)
	if !ok {
		return models.User{}, fmt.Errorf("invalid user ID in token claims")
	}
	id := uuid.MustParse(idStr)

	username, ok := userData["username"].(string)
	if !ok {
		return models.User{}, fmt.Errorf("invalid username in token claims")
	}

	email, ok := userData["email"].(string)
	if !ok {
		return models.User{}, fmt.Errorf("invalid email in token claims")
	}

	user := models.User{
		ID:       id,
		Username: username,
		Email:    email,
	}

	return user, nil
}
