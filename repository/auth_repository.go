package repository

import (
	"backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type AuthRepository interface {
	SaveRefreshToken(userID uint, tokenHash string, expiresAt time.Time) error
	DeleteRefreshToken(tokenHash string) error
	GetRefreshToken(tokenHash string) (*domain.RefreshToken, error)
	DeleteUserRefreshTokens(userID uint) error
	DeleteAllByUserID(userID uint) error 
}


type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) SaveRefreshToken(
	userID uint,
	tokenHash string,
	expiresAt time.Time,
) error {

	token := &domain.RefreshToken{
		UserID:     userID,
		Token:      tokenHash,
		ExpiresAt: expiresAt,
	}

	return r.db.Create(token).Error
}

func (r *authRepository) DeleteRefreshToken(tokenHash string) error {
	return r.db.
		Where("token = ?", tokenHash).
		Delete(&domain.RefreshToken{}).
		Error
}

func (r *authRepository) GetRefreshToken(tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken

	err := r.db.
		Where("token = ?", tokenHash).
		First(&token).
		Error

	return &token, err
}

func (r *authRepository) DeleteUserRefreshTokens(userID uint) error {
	return r.db.
		Where("user_id = ?", userID).
		Delete(&domain.RefreshToken{}).
		Error
}

func (r *authRepository) DeleteAllByUserID(userID uint) error {
	return r.db.
		Where("user_id = ?", userID).
		Delete(&domain.RefreshToken{}).
		Error
}
