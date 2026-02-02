package service

import (
	"backend/internal/domain"
	"backend/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

//GetProfile
func (s *UserService) GetProfile(userID uint) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	return user, nil
}


//Update
func (s *UserService) UpdateProfile(
	userID uint,
	newName string,
) error {

	if newName == "" {
		return ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if user.IsBlocked {
		return ErrUserBlocked
	}

	user.Name = newName

	return s.userRepo.Update(user)
}

//BlockUser

func (s *UserService) BlockUser(
	isAdmin bool,
	targetUserID uint,
) error {

	if !isAdmin {
		return ErrForbidden
	}

	user, err := s.userRepo.GetByID(targetUserID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if user.IsBlocked {
		return nil
	}

	user.IsBlocked = true

	return s.userRepo.Update(user)
}

//Admin

func (s *UserService) AdminUpdateUser(
	isAdmin bool,
	userID uint,
	name string,
	role string,
) error {

	if !isAdmin {
		return ErrForbidden
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if name != "" {
		user.Name = name
	}
	if role != "" {
		user.Role = role
	}

	return s.userRepo.Update(user)
}


