package handlers

import (
	"errors"
	"golang-order-manager-api/internal/dto"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	"golang-order-manager-api/internal/responses"
	"golang-order-manager-api/internal/services"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService services.UserService // interface, não struct concreta
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// UpdateUser godoc
//
// @Summary Update current user
// @Description Updates the authenticated user's profile information (email and/or username)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserRequest true "Update payload"
// @Success 200 {object} dto.UserResponse "Returns updated user data"
// @Failure 404 {object} responses.ErrorResponse "User not found"
// @Failure 409 {object} responses.ErrorResponse "Email or username already in use"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me [patch]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	req := dto.UpdateUserRequest{}

	c.Bind(&req)

	user := models.User{
		ID:       userID,
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	updatedUser, err := h.userService.Update(user)
	if err != nil {
		if errors.Is(err, error_codes.ErrEmailAlreadyInUse) {
			return responses.Error(c, http.StatusConflict, err)
		}

		if errors.Is(err, error_codes.ErrUsernameAlreadyInUse) {
			return responses.Error(c, http.StatusConflict, err)
		}

		if errors.Is(err, error_codes.ErrUserNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		slog.Error("Failed to update user", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	response := dto.UserResponse{
		ID:       updatedUser.ID,
		Username: updatedUser.Username,
		Email:    updatedUser.Email,
	}

	return responses.Success(c, http.StatusOK, "User updated successfully", response)
}

// DeleteUser godoc
//
// @Summary Delete current user
// @Description Soft-deletes the authenticated user's account
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 204 "User deleted successfully"
// @Failure 404 {object} responses.ErrorResponse "User not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	err := h.userService.Delete(userID)
	if err != nil {
		if errors.Is(err, error_codes.ErrUserNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		slog.Error("Failed to delete user", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return c.NoContent(http.StatusNoContent)
}

// GetUserInfo godoc
//
// @Summary Get current user info
// @Description Returns the authenticated user's profile information
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse "Returns user data"
// @Failure 404 {object} responses.ErrorResponse "User not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetUserInfo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		if errors.Is(err, error_codes.ErrUserNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		slog.Error("Failed to get user info", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	response := dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return responses.Success(c, http.StatusOK, "User info retrieved successfully", response)
}
