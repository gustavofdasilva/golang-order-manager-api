package handlers

import (
	"errors"
	"golang-order-manager-api/internal/config"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/responses"
	auth "golang-order-manager-api/internal/services"
	"golang-order-manager-api/pkg/database"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Register(c echo.Context) error {
	user := models.User{}

	c.Bind(&user)

	if user.Email == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidEmail)
	}
	if user.Password == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidPassword)
	}
	if user.Username == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidUsername)
	}

	db := database.GetDB()

	repo := repository.NewUserRepo(db)

	authService := auth.NewAuthService(&repo, nil, config.SECRET_KEY, time.Duration(config.REFRESH_TOKEN_EXPIRATION_MINUTES)*time.Minute)

	_, err := authService.Register(user.Username, user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to register user", slog.Any("err", err))

		if errors.Is(err, error_codes.ErrEmailAlreadyInUse) {
			return responses.Error(c, http.StatusBadRequest, err)
		}

		if errors.Is(err, error_codes.ErrUsernameAlreadyInUse) {
			return responses.Error(c, http.StatusBadRequest, err)
		}

		slog.Error("Failed to register user", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return c.NoContent(http.StatusNoContent)
}

func Login(c echo.Context) error {
	user := models.User{}

	c.Bind(&user)

	if user.Email == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidEmail)
	}
	if user.Password == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidPassword)
	}

	db := database.GetDB()

	repoUser := repository.NewUserRepo(db)
	repoAuth := repository.NewAuthRepo(db)

	authService := auth.NewAuthService(&repoUser, &repoAuth, config.SECRET_KEY, time.Duration(config.REFRESH_TOKEN_EXPIRATION_MINUTES)*time.Minute)

	token, refreshToken, user, err := authService.Login(user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to login user", slog.Any("err", err))

		if errors.Is(err, error_codes.ErrInvalidCredentials) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		if errors.Is(err, error_codes.ErrUserNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		slog.Error("Failed to login user", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	response := make(map[string]interface{})
	response["access_token"] = token
	response["refresh_token"] = refreshToken
	response["user"] = user

	return responses.Success(c, http.StatusOK, "Login successful", response)
}

func RefreshToken(c echo.Context) error {
	body := struct {
		RefreshToken string `json:"refresh_token"`
	}{}

	c.Bind(&body)

	if body.RefreshToken == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidRefreshToken)
	}

	db := database.GetDB()

	repoUser := repository.NewUserRepo(db)
	repoAuth := repository.NewAuthRepo(db)

	authService := auth.NewAuthService(&repoUser, &repoAuth, config.SECRET_KEY, time.Duration(config.REFRESH_TOKEN_EXPIRATION_MINUTES)*time.Minute)

	token, refreshToken, user, err := authService.Refresh(body.RefreshToken)
	if err != nil {

		if errors.Is(err, error_codes.ErrRefreshTokenNotFound) || errors.Is(err, error_codes.ErrInvalidRefreshToken) || errors.Is(err, error_codes.ErrRefreshTokenExpired) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		slog.Error("Failed to refresh token", slog.Any("err", err))
		return responses.Error(c, http.StatusUnauthorized, err)
	}

	response := make(map[string]interface{})
	response["access_token"] = token
	response["refresh_token"] = refreshToken
	response["user"] = user

	return responses.Success(c, http.StatusOK, "Refresh successful", response)
}

func Logout(c echo.Context) error {
	body := struct {
		RefreshToken string `json:"refresh_token"`
	}{}

	c.Bind(&body)

	if body.RefreshToken == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidRefreshToken)
	}

	db := database.GetDB()

	repoUser := repository.NewUserRepo(db)
	repoAuth := repository.NewAuthRepo(db)

	authService := auth.NewAuthService(&repoUser, &repoAuth, config.SECRET_KEY, time.Duration(config.REFRESH_TOKEN_EXPIRATION_MINUTES)*time.Minute)

	err := authService.Logout(body.RefreshToken)
	if err != nil {

		if errors.Is(err, error_codes.ErrRefreshTokenNotFound) || errors.Is(err, error_codes.ErrInvalidRefreshToken) || errors.Is(err, error_codes.ErrRefreshTokenExpired) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		slog.Error("Failed to logout user", slog.Any("err", err))
		return responses.Error(c, http.StatusUnauthorized, err)
	}

	return responses.Success(c, http.StatusOK, "Logout successful", nil)
}

func LogoutAll(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	db := database.GetDB()

	repoUser := repository.NewUserRepo(db)
	repoAuth := repository.NewAuthRepo(db)

	authService := auth.NewAuthService(&repoUser, &repoAuth, config.SECRET_KEY, time.Duration(config.REFRESH_TOKEN_EXPIRATION_MINUTES)*time.Minute)

	err := authService.LogoutAll(userID)
	if err != nil {

		if errors.Is(err, error_codes.ErrRefreshTokenNotFound) || errors.Is(err, error_codes.ErrInvalidRefreshToken) || errors.Is(err, error_codes.ErrRefreshTokenExpired) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		slog.Error("Failed to logout user", slog.Any("err", err))
		return responses.Error(c, http.StatusUnauthorized, err)
	}

	return responses.Success(c, http.StatusOK, "Logout successful", nil)
}
