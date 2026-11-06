package services

import (
	error_codes "golang-order-manager-api/internal/errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/security"

	"github.com/google/uuid"
)

type UserService interface {
	Update(user models.User) (models.User, error)
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Update(user models.User) (models.User, error) {

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

func (s *userService) Delete(id uuid.UUID) error {

	err := s.userRepo.DeleteUserByID(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) GetByID(id uuid.UUID) (models.User, error) {

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return models.User{}, err
	}

	user.Password = ""

	return user, nil
}
