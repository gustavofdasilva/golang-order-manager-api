package handlers

import (
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/services"
	"golang-order-manager-api/pkg/database"
	"log"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func UpdateUser(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	user := models.User{}

	c.Bind(&user)
	user.ID = userID
	log.Println(userID)

	db := database.GetDB()

	repo := repository.NewUserRepo(db)
	userService := services.NewUserService(&repo)

	newUser, err := userService.Update(user)
	if err != nil {
		slog.Error("Failed to update user", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(http.StatusOK, newUser)
}

func DeleteUser(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	db := database.GetDB()

	repo := repository.NewUserRepo(db)
	userService := services.NewUserService(&repo)

	err := userService.Delete(userID)
	if err != nil {
		slog.Error("Failed to delete user", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to delete user")
	}

	return c.JSON(http.StatusOK, "User deleted successfully")
}

func GetUserInfo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	db := database.GetDB()

	repoUser := repository.NewUserRepo(db)

	userService := services.NewUserService(&repoUser)

	user, err := userService.GetByID(userID)
	if err != nil {
		slog.Error("Failed to get user info", slog.Any("err", err))
		return c.JSON(http.StatusInternalServerError, "Failed to get user info")
	}

	return c.JSON(http.StatusOK, user)
}
