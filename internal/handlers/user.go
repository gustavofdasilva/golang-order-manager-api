package handlers

import (
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/services"
	"golang-order-manager-api/pkg/database"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func CreateUser(c echo.Context) error {
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

	newUser, err := repo.CreateUser(user.Username, user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to create user", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to create user")
	}

	token, err := services.CreateJWTToken(newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(http.StatusOK, token)
}

func DeleteUser(c echo.Context) error {
	user := models.User{}

	c.Bind(&user)

	db := database.GetDB()

	repo := repository.NewUserRepo(db)

	err := repo.DeleteUser(user.ID)
	if err != nil {
		slog.Error("Failed to delete user", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to delete user")
	}

	return c.JSON(http.StatusOK, "User deleted successfully")
}

func UpdateUser(c echo.Context) error {
	user := models.User{}

	c.Bind(&user)

	db := database.GetDB()

	repo := repository.NewUserRepo(db)

	err := repo.UpdateUser(user)
	if err != nil {
		slog.Error("Failed to update user", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(http.StatusOK, "User updated successfully")
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

	repoAuth := repository.NewAuthRepo(db)
	repoUser := repository.NewUserRepo(db)

	emailAlreadyInUse, err := repoUser.IsEmailAlreadyInUse(user.Email)
	if err != nil {
		slog.Error("Failed to check email", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to check email")
	}

	if !emailAlreadyInUse {
		return c.JSON(http.StatusNotFound, "User not found")
	}

	userInfo, err := repoAuth.LoginUser(user.Email, user.Password)
	if err != nil {
		slog.Error("Failed to login user", slog.Any("err", err))
		return c.JSON(http.StatusUnauthorized, "Invalid email or password")
	}

	token, err := services.CreateJWTToken(userInfo)
	if err != nil {
		log.Fatalf("error to create token: %v", err)
	}

	return c.JSON(http.StatusOK, token)
}

func GetUserInfo(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")

	if authHeader == "" {
		return c.JSON((http.StatusUnauthorized), "missing authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return c.JSON((http.StatusUnauthorized), "invalid token format")
	}

	token := strings.TrimPrefix(authHeader, prefix)

	user, err := services.ReadTokenToUser(token)
	if err != nil {
		slog.Error("Failed to read token", slog.Any("err", err))
		return c.JSON(http.StatusUnauthorized, "invalid token")
	}

	return c.JSON(http.StatusOK, user)
}
