package service

import "backend/internal/domain"

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
	//  REVOKE ALL SESSIONS
	if err := s.authRepo.DeleteAllByUserID(targetUserID); err != nil {
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

func (s *UserService) ListAllWithCount(page, limit int) ([]domain.User, int64, error) {
	offset := (page - 1) * limit

	users, err := s.userRepo.List(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count()
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
