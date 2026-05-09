package handlers

import (
	"golang-order-manager-api/internal/config"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
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
		return c.JSON(http.StatusBadRequest, "Invalid email")
	}
	if user.Password == "" {
		return c.JSON(http.StatusBadRequest, "Invalid password")
	}
	if user.Username == "" {
		return c.JSON(http.StatusBadRequest, "Invalid username")
	}

	db := database.GetDB()

	repo := repository.NewUserRepo(db)

	authService := auth.NewAuthService(&repo, config.SECRET_KEY)

	_, err := authService.Register(user.Username, user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to register user", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to create user")
	}

	return c.NoContent(http.StatusNoContent)
}

func Login(c echo.Context) error {
	user := models.User{}

	c.Bind(&user)

	if user.Email == "" {
		return c.JSON(http.StatusBadRequest, "Invalid email")
	}
	if user.Password == "" {
		return c.JSON(http.StatusBadRequest, "Invalid password")
	}

	db := database.GetDB()

	repoUser := repository.NewUserRepo(db)

	authService := auth.NewAuthService(&repoUser, config.SECRET_KEY)

	//? TODO: refactor to return only token, not user
	token, _, err := authService.Login(user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to login user", slog.Any("err", err))
		return c.JSON(http.StatusUnauthorized, "Invalid email or password")
	}

	response := make(map[string]interface{})
	response["token"] = token

	return c.JSON(http.StatusOK, response)
}
