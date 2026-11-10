package handlers

import (
	"errors"
	"golang-order-manager-api/internal/dto"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/responses"
	"golang-order-manager-api/internal/services"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
//
// @Summary Register new user
// @Description Creates a new user account. Requires email, username, and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration payload"
// @Success 204 "User created successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid fields or email/username already in use"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	req := dto.RegisterRequest{}

	c.Bind(&req)

	if req.Email == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidEmail)
	}
	if req.Password == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidPassword)
	}
	if req.Username == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidUsername)
	}

	_, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
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

// Login godoc
//
// @Summary Login user
// @Description Authenticates a user and returns access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.TokensResponse "Returns access_token, refresh_token, and user info"
// @Failure 400 {object} responses.ErrorResponse "Missing email or password"
// @Failure 401 {object} responses.ErrorResponse "Invalid credentials"
// @Failure 404 {object} responses.ErrorResponse "User not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	req := dto.LoginRequest{}

	c.Bind(&req)

	if req.Email == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidEmail)
	}
	if req.Password == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidPassword)
	}

	token, refreshToken, user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, error_codes.ErrInvalidCredentials) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		if errors.Is(err, error_codes.ErrUserNotFound) {
			return responses.Error(c, http.StatusNotFound, err)
		}

		slog.Error("Failed to login user", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	response := dto.TokensResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}

	return responses.Success(c, http.StatusOK, "Login successful", response)
}

// RefreshToken godoc
//
// @Summary Refresh access token
// @Description Rotates the access and refresh token pair using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} dto.TokensResponse "Returns new access_token, refresh_token, and user info"
// @Failure 400 {object} responses.ErrorResponse "Missing refresh token"
// @Failure 401 {object} responses.ErrorResponse "Invalid, expired, or not found refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	req := dto.RefreshTokenRequest{}

	c.Bind(&req)

	if req.RefreshToken == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidRefreshToken)
	}

	token, newRefreshToken, user, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, error_codes.ErrRefreshTokenNotFound) || errors.Is(err, error_codes.ErrInvalidRefreshToken) || errors.Is(err, error_codes.ErrRefreshTokenExpired) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		slog.Error("Failed to refresh token", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	response := dto.TokensResponse{
		AccessToken:  token,
		RefreshToken: newRefreshToken,
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}

	return responses.Success(c, http.StatusOK, "Refresh successful", response)
}

// Logout godoc
//
// @Summary Logout user
// @Description Invalidates a specific refresh token, ending the current session
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RefreshTokenRequest true "Refresh token to invalidate"
// @Success 200 {object} responses.SuccessResponse "Logout successful"
// @Failure 400 {object} responses.ErrorResponse "Missing refresh token"
// @Failure 401 {object} responses.ErrorResponse "Invalid, expired, or not found refresh token"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	req := dto.RefreshTokenRequest{}

	c.Bind(&req)

	if req.RefreshToken == "" {
		return responses.Error(c, http.StatusBadRequest, error_codes.ErrInvalidRefreshToken)
	}

	err := h.authService.Logout(req.RefreshToken)
	if err != nil {
		if errors.Is(err, error_codes.ErrRefreshTokenNotFound) || errors.Is(err, error_codes.ErrInvalidRefreshToken) || errors.Is(err, error_codes.ErrRefreshTokenExpired) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		slog.Error("Failed to logout user", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return responses.Success(c, http.StatusOK, "Logout successful", nil)
}

// LogoutAll godoc
//
// @Summary Logout from all devices
// @Description Invalidates all refresh tokens for the authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.SuccessResponse "Logout successful"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized or invalid session"
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	err := h.authService.LogoutAll(userID)
	if err != nil {
		if errors.Is(err, error_codes.ErrRefreshTokenNotFound) || errors.Is(err, error_codes.ErrInvalidRefreshToken) || errors.Is(err, error_codes.ErrRefreshTokenExpired) {
			return responses.Error(c, http.StatusUnauthorized, err)
		}

		slog.Error("Failed to logout all sessions", slog.Any("err", err))
		return responses.Error(c, http.StatusInternalServerError, error_codes.ErrUnexpectedError)
	}

	return responses.Success(c, http.StatusOK, "Logout successful", nil)
}
