package services

import (
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/security"
	"log/slog"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Update(user models.User) (models.User, error) {

	exists, err := s.userRepo.IsEmailAlreadyInUse(user.Email, &user.ID)
	if err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, error_codes.ErrEmailAlreadyInUse
	}

	exists, err = s.userRepo.IsUsernameAlreadyInUse(user.Username, &user.ID)
	if err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, error_codes.ErrUsernameAlreadyInUse
	}

	hashedPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return models.User{}, err
	}

	user.Password = hashedPassword

	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return models.User{}, err
	}

	user.Password = ""

	return user, nil
}

func (s *UserService) Delete(id uuid.UUID) error {

	err := s.userRepo.DeleteUserByID(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetByID(id uuid.UUID) (models.User, error) {

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		slog.Error("Failed to get user by ID", slog.Any("err", err), slog.Any("userID", id))
		return models.User{}, err
	}

	user.Password = ""

	return user, nil
}
