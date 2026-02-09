package repository

import (
	"backend/internal/domain"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

type userRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// Constructor
func NewUserRepository(db *gorm.DB, logger *slog.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// -----------------------------------------------------

func (r *userRepository) Create(user *domain.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		r.logger.Error(
			"user create failed",
			"email", user.Email,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"user created",
		"user_id", user.ID,
		"email", user.Email,
	)
	return nil
}

func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User

	err := r.db.First(&user, id).Error
	if err != nil {
		r.logger.Error(
			"user get by id failed",
			"user_id", id,
			"err", err,
		)
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User

	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// NOT an error condition
			r.logger.Info(
				"user not found by email",
				"email", email,
			)
			return nil, nil
		}

		r.logger.Error(
			"user get by email failed",
			"email", email,
			"err", err,
		)
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	err := r.db.Save(user).Error
	if err != nil {
		r.logger.Error(
			"user update failed",
			"user_id", user.ID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"user updated",
		"user_id", user.ID,
	)
	return nil
}

func (r *userRepository) UpdatePassword(userID uint, hashedPassword string) error {
	err := r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).
		Error

	if err != nil {
		r.logger.Error(
			"user password update failed",
			"user_id", userID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"user password updated",
		"user_id", userID,
	)
	return nil
}

func (r *userRepository) Delete(id uint) error {
	err := r.db.Delete(&domain.User{}, id).Error
	if err != nil {
		r.logger.Error(
			"user delete failed",
			"user_id", id,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"user deleted",
		"user_id", id,
	)
	return nil
}

func (r *userRepository) List(offset, limit int) ([]domain.User, error) {
	var users []domain.User

	query := r.db.Model(&domain.User{}).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// Count returns total number of users
func (r *userRepository) Count() (int64, error) {
	var total int64

	if err := r.db.Model(&domain.User{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}