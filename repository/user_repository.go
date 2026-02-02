package repository

import (
	"backend/internal/domain"
	"errors"

	"gorm.io/gorm"
)

//----------------------------------------------------------------------------------

type userRepository struct {
	db *gorm.DB
}

/*It stores the DB connection (*gorm.DB) once.
It allows methods like Create, GetByID, etc. to reuse that same DB connection via r.db.
It acts as the concrete implementation of the UserRepository interface.
-------------------------------------------------------------------------------*/

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db} //it links the DB connection to the repository so the repository can interact with the database.
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for "not found"
		}
		return nil, err // Return actual error
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdatePassword(userID uint, hashedPassword string) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).
		Error
}


func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

/*
What this file actually is

This file is a concrete repository implementation.
It answers HOW data is stored and retrieved, using:

GORM
A relational database

*/
