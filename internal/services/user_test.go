package services_test

import (
	"fmt"
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repomocks "golang-order-manager-api/internal/repositories/mocks"
	"golang-order-manager-api/internal/services"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeUserService(t *testing.T) (*services.UserService, *repomocks.MockUserRepository) {
	mockRepo := repomocks.NewMockUserRepository(t)
	svc := services.NewUserService(mockRepo)
	return svc, mockRepo
}

func TestGetByID_Success(t *testing.T) {
	svc, mockRepo := makeUserService(t)
	dummyID := uuid.New()

	expected := models.User{
		ID:       dummyID,
		Email:    "john@example.com",
		Username: "john",
	}

	mockRepo.EXPECT().
		GetByID(dummyID).
		Return(expected, nil)

	result, err := svc.GetByID(dummyID)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Username, result.Username)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()

	mockRepo.EXPECT().
		GetByID(dummyID).
		Return(models.User{}, error_codes.ErrUserNotFound)

	_, err := svc.GetByID(dummyID)

	assert.ErrorIs(t, err, error_codes.ErrUserNotFound)
}

func TestUpdate_Success(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()
	input := models.User{
		ID:       dummyID,
		Email:    "john@example.com",
		Username: "john",
		Password: "plaintext123",
	}

	mockRepo.EXPECT().
		IsEmailAlreadyInUse(input.Email, &dummyID).
		Return(false, nil)

	mockRepo.EXPECT().
		IsUsernameAlreadyInUse(input.Username, &dummyID).
		Return(false, nil)

	mockRepo.EXPECT().
		UpdateUser(mock.MatchedBy(func(u models.User) bool {
			// senha deve ter sido hasheada antes de chegar no repo
			return u.Password != "plaintext123" && u.Email == input.Email
		})).
		Return(nil)

	result, err := svc.Update(input)

	require.NoError(t, err)
	assert.Equal(t, input.Email, result.Email)
	assert.Empty(t, result.Password) // service limpa a senha antes de retornar
}

func TestUpdate_EmailAlreadyInUse(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()
	input := models.User{ID: dummyID, Email: "taken@example.com", Username: "john"}

	mockRepo.EXPECT().
		IsEmailAlreadyInUse(input.Email, &dummyID).
		Return(true, nil)

	_, err := svc.Update(input)

	assert.ErrorIs(t, err, error_codes.ErrEmailAlreadyInUse)
}

func TestUpdate_UsernameAlreadyInUse(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()
	input := models.User{ID: dummyID, Email: "john@example.com", Username: "taken"}

	mockRepo.EXPECT().
		IsEmailAlreadyInUse(input.Email, &dummyID).
		Return(false, nil)

	mockRepo.EXPECT().
		IsUsernameAlreadyInUse(input.Username, &dummyID).
		Return(true, nil)

	_, err := svc.Update(input)

	assert.ErrorIs(t, err, error_codes.ErrUsernameAlreadyInUse)
}

func TestUpdate_RepoError(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()
	input := models.User{ID: dummyID, Email: "john@example.com", Username: "john", Password: "123"}

	mockRepo.EXPECT().IsEmailAlreadyInUse(input.Email, &dummyID).Return(false, nil)
	mockRepo.EXPECT().IsUsernameAlreadyInUse(input.Username, &dummyID).Return(false, nil)
	mockRepo.EXPECT().UpdateUser(mock.Anything).Return(fmt.Errorf("db connection lost"))

	_, err := svc.Update(input)

	assert.Error(t, err)
}

func TestDelete_Success(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()

	mockRepo.EXPECT().DeleteUserByID(dummyID).Return(nil)

	err := svc.Delete(dummyID)

	assert.NoError(t, err)
}

func TestDelete_UserNotFound(t *testing.T) {
	svc, mockRepo := makeUserService(t)

	dummyID := uuid.New()

	mockRepo.EXPECT().DeleteUserByID(dummyID).Return(error_codes.ErrUserNotFound)

	err := svc.Delete(dummyID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, error_codes.ErrUserNotFound)
}
