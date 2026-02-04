package service

import (
	"backend/internal/domain"
	"backend/repository"
	"log/slog"
)

type UserService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	logger *slog.Logger,
) *UserService {
	if logger == nil {
		panic("UserService requires a non-nil logger")
	}

	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetProfile retrieves a user profile by ID
func (s *UserService) GetProfile(userID uint) (*domain.User, error) {
	if userID == 0 {
		s.logger.Warn("invalid userID", "user_id", userID)
		return nil, ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		s.logger.Error("user not found", "user_id", userID, "err", err)
		return nil, ErrUserNotFound
	}

	if user.IsBlocked {
		s.logger.Warn("blocked user access", "user_id", userID)
		return nil, ErrUserBlocked
	}

	return user, nil
}

// UpdateProfile updates the name of a user
func (s *UserService) UpdateProfile(userID uint, newName string) error {
	if userID == 0 || newName == "" {
		s.logger.Warn("invalid input", "user_id", userID, "name", newName)
		return ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		s.logger.Error("user not found", "user_id", userID, "err", err)
		return ErrUserNotFound
	}

	if user.IsBlocked {
		s.logger.Warn("blocked user update attempt", "user_id", userID)
		return ErrUserBlocked
	}

	user.Name = newName

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("db update failed", "user_id", userID, "err", err)
		return err
	}

	return nil
}

// BlockUser marks a user as blocked (admin only)
func (s *UserService) BlockUser(isAdmin bool, targetUserID uint) error {
	if !isAdmin {
		s.logger.Warn("non-admin block attempt", "target_user_id", targetUserID)
		return ErrForbidden
	}

	if targetUserID == 0 {
		s.logger.Warn("invalid targetUserID", "target_user_id", targetUserID)
		return ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(targetUserID)
	if err != nil || user == nil {
		s.logger.Error("user not found", "target_user_id", targetUserID, "err", err)
		return ErrUserNotFound
	}

	if user.IsBlocked {
		s.logger.Info("user already blocked", "target_user_id", targetUserID)
		return nil
	}

	user.IsBlocked = true

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("db update failed", "target_user_id", targetUserID, "err", err)
		return err
	}

	s.logger.Info("user blocked successfully", "target_user_id", targetUserID)
	return nil
}

// AdminUpdateUser updates user details (admin only)
func (s *UserService) AdminUpdateUser(isAdmin bool, userID uint, name string, role string) error {
	if !isAdmin {
		s.logger.Warn("non-admin update attempt", "user_id", userID)
		return ErrForbidden
	}

	if userID == 0 {
		s.logger.Warn("invalid userID", "user_id", userID)
		return ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		s.logger.Error("user not found", "user_id", userID, "err", err)
		return ErrUserNotFound
	}

	if name != "" {
		user.Name = name
	}

	if role != "" {
		user.Role = role
	}

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("db update failed", "user_id", userID, "err", err)
		return err
	}

	s.logger.Info("admin user update success", "user_id", userID, "name", name, "role", role)
	return nil
}
