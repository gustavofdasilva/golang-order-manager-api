package handlers_test

import (
	"encoding/json"
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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- setup ---

type userHandlerMocks struct {
	userService *svcmocks.MockUserService
}

func makeUserHandler(t *testing.T) (*handlers.UserHandler, *userHandlerMocks) {
	m := &userHandlerMocks{
		userService: svcmocks.NewMockUserService(t),
	}
	h := handlers.NewUserHandler(m.userService)
	return h, m
}

// helper: monta contexto Echo com userID já injetado (simula middleware de auth)
func newEchoContext(method, path, body string, userID uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
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
	c.Set("userID", userID)
	return c, rec
}

// helper: deserializa o body da resposta
func parseBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	return body
}

// --- GetUserInfo ---

func TestGetUserInfo_Success(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	expected := models.User{ID: userID, Email: "john@example.com", Username: "john"}

	m.userService.EXPECT().GetByID(userID).Return(expected, nil)

	c, rec := newEchoContext(http.MethodGet, "/users/me", "", userID)
	err := h.GetUserInfo(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := parseBody(t, rec)
	data := body["data"].(map[string]any)
	assert.Equal(t, "john@example.com", data["email"])
	assert.Equal(t, "john", data["username"])
}

func TestGetUserInfo_NotFound(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().GetByID(userID).Return(models.User{}, error_codes.ErrUserNotFound)

	c, rec := newEchoContext(http.MethodGet, "/users/me", "", userID)
	err := h.GetUserInfo(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetUserInfo_InternalError(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().GetByID(userID).Return(models.User{}, fmt.Errorf("db connection lost"))

	c, rec := newEchoContext(http.MethodGet, "/users/me", "", userID)
	err := h.GetUserInfo(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- UpdateUser ---

func TestUpdateUser_Success(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	body := `{"email":"new@example.com","username":"newname","password":"newpass"}`

	m.userService.EXPECT().
		Update(models.User{
			ID:       userID,
			Email:    "new@example.com",
			Username: "newname",
			Password: "newpass",
		}).
		Return(models.User{ID: userID, Email: "new@example.com", Username: "newname"}, nil)

	c, rec := newEchoContext(http.MethodPatch, "/users/me", body, userID)
	err := h.UpdateUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	data := parseBody(t, rec)["data"].(map[string]any)
	assert.Equal(t, "new@example.com", data["email"])
}

func TestUpdateUser_EmailAlreadyInUse(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	body := `{"email":"taken@example.com","username":"john","password":"123"}`

	m.userService.EXPECT().
		Update(mock.Anything).
		Return(models.User{}, error_codes.ErrEmailAlreadyInUse)

	c, rec := newEchoContext(http.MethodPatch, "/users/me", body, userID)
	err := h.UpdateUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestUpdateUser_UsernameAlreadyInUse(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	body := `{"email":"john@example.com","username":"taken","password":"123"}`

	m.userService.EXPECT().
		Update(mock.Anything).
		Return(models.User{}, error_codes.ErrUsernameAlreadyInUse)

	c, rec := newEchoContext(http.MethodPatch, "/users/me", body, userID)
	err := h.UpdateUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestUpdateUser_NotFound(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().Update(mock.Anything).Return(models.User{}, error_codes.ErrUserNotFound)

	c, rec := newEchoContext(http.MethodPatch, "/users/me", `{}`, userID)
	err := h.UpdateUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateUser_InternalError(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().Update(mock.Anything).Return(models.User{}, fmt.Errorf("db connection lost"))

	c, rec := newEchoContext(http.MethodPatch, "/users/me", `{}`, userID)
	err := h.UpdateUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- DeleteUser ---

func TestDeleteUser_Success(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().Delete(userID).Return(nil)

	c, rec := newEchoContext(http.MethodDelete, "/users/me", "", userID)
	err := h.DeleteUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteUser_NotFound(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().Delete(userID).Return(error_codes.ErrUserNotFound)

	c, rec := newEchoContext(http.MethodDelete, "/users/me", "", userID)
	err := h.DeleteUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteUser_InternalError(t *testing.T) {
	h, m := makeUserHandler(t)

	userID := uuid.New()
	m.userService.EXPECT().Delete(userID).Return(fmt.Errorf("db connection lost"))

	c, rec := newEchoContext(http.MethodDelete, "/users/me", "", userID)
	err := h.DeleteUser(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
