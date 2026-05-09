package services

import (
	"errors"
	"golang-order-manager-api/internal/models"
	repository "golang-order-manager-api/internal/repositories"
	"golang-order-manager-api/internal/security"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
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
		return models.User{}, errors.New("new email already in use")
	}

	exists, err = s.userRepo.IsUsernameAlreadyInUse(user.Username, &user.ID)
	if err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, errors.New("new username already in use")
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
		return models.User{}, err
	}

	return user, nil
}
