package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/handlers"
	"golang-order-manager-api/internal/models"
	svcmocks "golang-order-manager-api/internal/services/mocks"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- setup ---

func makeAuthHandler(t *testing.T) (*handlers.AuthHandler, *svcmocks.MockAuthService) {
	mockSvc := svcmocks.NewMockAuthService(t)
	h := handlers.NewAuthHandler(mockSvc)
	return h, mockSvc
}

func newAuthContext(method, path, body string, userID *uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if userID != nil {
		c.Set("userID", *userID)
	}
	return c, rec
}

func fakeUser() models.User {
	return models.User{
		ID:       uuid.New(),
		Email:    "john@example.com",
		Username: "john",
	}
}

func fakeUserLogin() models.User {
	return models.User{
		Email:    "john@example.com",
		Username: "john",
		Password: "secret123",
	}
}

func TestRegister_Success(t *testing.T) {
	h, m := makeAuthHandler(t)

	body := toJSON(t, fakeUserLogin())

	m.EXPECT().
		Register("john", "john@example.com", "secret123").
		Return(fakeUser(), nil)

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRegister_MissingEmail(t *testing.T) {
	h, _ := makeAuthHandler(t)

	user := fakeUserLogin()
	user.Email = ""
	body := toJSON(t, user)

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_MissingPassword(t *testing.T) {
	h, _ := makeAuthHandler(t)

	user := fakeUserLogin()
	user.Password = ""
	body := toJSON(t, user)

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_MissingUsername(t *testing.T) {
	h, _ := makeAuthHandler(t)

	user := fakeUserLogin()
	user.Username = ""
	body := toJSON(t, user)

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_EmailAlreadyInUse(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUserLogin()
	body := toJSON(t, user)

	m.EXPECT().
		Register(user.Username, user.Email, user.Password).
		Return(models.User{}, error_codes.ErrEmailAlreadyInUse)

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_UsernameAlreadyInUse(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUserLogin()
	body := toJSON(t, user)

	m.EXPECT().
		Register(user.Username, user.Email, user.Password).
		Return(models.User{}, error_codes.ErrUsernameAlreadyInUse)

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_InternalError(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUserLogin()
	body := toJSON(t, user)

	m.EXPECT().
		Register(user.Username, user.Email, user.Password).
		Return(models.User{}, fmt.Errorf("unexpected db error"))

	c, rec := newAuthContext(http.MethodPost, "/auth/register", body, nil)
	err := h.Register(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestLogin_Success(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUser()
	userReq := fakeUserLogin()
	body := toJSON(t, userReq)
	m.EXPECT().
		Login(userReq.Email, userReq.Password).
		Return("access-token", "refresh-token", user, nil)

	c, rec := newAuthContext(http.MethodPost, "/auth/login", body, nil)
	err := h.Login(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, "access-token", data["access_token"])
	assert.Equal(t, "refresh-token", data["refresh_token"])
	userResp := data["user"].(map[string]any)
	assert.Equal(t, "john@example.com", userResp["email"])
}

func TestLogin_MissingEmail(t *testing.T) {
	h, _ := makeAuthHandler(t)

	user := fakeUserLogin()
	user.Email = ""
	body := toJSON(t, user)

	c, rec := newAuthContext(http.MethodPost, "/auth/login", body, nil)
	err := h.Login(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_MissingPassword(t *testing.T) {
	h, _ := makeAuthHandler(t)

	user := fakeUserLogin()
	user.Password = ""
	body := toJSON(t, user)

	c, rec := newAuthContext(http.MethodPost, "/auth/login", body, nil)
	err := h.Login(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUserLogin()
	body := toJSON(t, user)

	m.EXPECT().
		Login(user.Email, user.Password).
		Return("", "", models.User{}, error_codes.ErrInvalidCredentials)

	c, rec := newAuthContext(http.MethodPost, "/auth/login", body, nil)
	err := h.Login(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUserLogin()
	body := toJSON(t, user)

	m.EXPECT().
		Login(user.Email, user.Password).
		Return("", "", models.User{}, error_codes.ErrUserNotFound)

	c, rec := newAuthContext(http.MethodPost, "/auth/login", body, nil)
	err := h.Login(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRefreshToken_Success(t *testing.T) {
	h, m := makeAuthHandler(t)

	user := fakeUser()
	body := `{"refresh_token":"old-refresh-token"}`

	m.EXPECT().
		Refresh("old-refresh-token").
		Return("new-access-token", "new-refresh-token", user, nil)

	c, rec := newAuthContext(http.MethodPost, "/auth/refresh", body, nil)
	err := h.RefreshToken(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, "new-access-token", data["access_token"])
	assert.Equal(t, "new-refresh-token", data["refresh_token"])
}

func TestRefreshToken_MissingToken(t *testing.T) {
	h, _ := makeAuthHandler(t)

	c, rec := newAuthContext(http.MethodPost, "/auth/refresh", `{}`, nil)
	err := h.RefreshToken(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRefreshToken_Expired(t *testing.T) {
	h, m := makeAuthHandler(t)

	m.EXPECT().
		Refresh("expired-token").
		Return("", "", models.User{}, error_codes.ErrRefreshTokenExpired)

	body := `{"refresh_token":"expired-token"}`
	c, rec := newAuthContext(http.MethodPost, "/auth/refresh", body, nil)
	err := h.RefreshToken(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRefreshToken_NotFound(t *testing.T) {
	h, m := makeAuthHandler(t)

	m.EXPECT().
		Refresh("unknown-token").
		Return("", "", models.User{}, error_codes.ErrRefreshTokenNotFound)

	body := `{"refresh_token":"unknown-token"}`
	c, rec := newAuthContext(http.MethodPost, "/auth/refresh", body, nil)
	err := h.RefreshToken(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogout_Success(t *testing.T) {
	h, m := makeAuthHandler(t)

	m.EXPECT().Logout("valid-token").Return(nil)

	body := `{"refresh_token":"valid-token"}`
	c, rec := newAuthContext(http.MethodPost, "/auth/logout", body, nil)
	err := h.Logout(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLogout_MissingToken(t *testing.T) {
	h, _ := makeAuthHandler(t)

	c, rec := newAuthContext(http.MethodPost, "/auth/logout", `{}`, nil)
	err := h.Logout(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogout_ExpiredToken(t *testing.T) {
	h, m := makeAuthHandler(t)

	m.EXPECT().Logout("expired-token").Return(error_codes.ErrRefreshTokenExpired)

	body := `{"refresh_token":"expired-token"}`
	c, rec := newAuthContext(http.MethodPost, "/auth/logout", body, nil)
	err := h.Logout(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogout_InternalError(t *testing.T) {
	h, m := makeAuthHandler(t)

	m.EXPECT().Logout("valid-token").Return(fmt.Errorf("db error"))

	body := `{"refresh_token":"valid-token"}`
	c, rec := newAuthContext(http.MethodPost, "/auth/logout", body, nil)
	err := h.Logout(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestLogoutAll_Success(t *testing.T) {
	h, m := makeAuthHandler(t)

	userID := uuid.New()
	m.EXPECT().LogoutAll(userID).Return(nil)

	c, rec := newAuthContext(http.MethodPost, "/auth/logout-all", "", &userID)
	err := h.LogoutAll(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLogoutAll_InternalError(t *testing.T) {
	h, mockSvc := makeAuthHandler(t)

	userID := uuid.New()
	mockSvc.EXPECT().LogoutAll(userID).Return(fmt.Errorf("db error"))

	c, rec := newAuthContext(http.MethodPost, "/auth/logout-all", "", &userID)
	err := h.LogoutAll(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
