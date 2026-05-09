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

	authService := auth.NewAuthService(&repo, config.SECRET_KEY)

	_, err := authService.Register(user.Username, user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to register user", slog.Any("err", err))

		if errors.Is(err, error_codes.ErrEmailAlreadyInUse) {
			return responses.Error(c, http.StatusBadRequest, err)
		}

		if errors.Is(err, error_codes.ErrUsernameAlreadyInUse) {
			return responses.Error(c, http.StatusBadRequest, err)
		}

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

	authService := auth.NewAuthService(&repoUser, config.SECRET_KEY)

	//? TODO: refactor to return only token, not user
	token, user, err := authService.Login(user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to login user", slog.Any("err", err))

		if errors.Is(err, error_codes.ErrInvalidCredentials) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		if errors.Is(err, error_codes.ErrUserNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	response := make(map[string]interface{})
	response["access_token"] = token
	response["user"] = user

	return responses.Success(c, http.StatusOK, "Login successful", response)
}
