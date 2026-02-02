package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type WishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) *WishlistRepository{
	return &WishlistRepository{db:db}
}

func(r *WishlistRepository) Add(item *domain.WishlistItem) error{
  return r.db.Create(&item).Error
}

func(r *WishlistRepository) Remove(UserID,ProductID uint) error{
  return r.db.Where("user_id=? AND product_id=?",UserID,ProductID).
  Delete(&domain.WishlistItem{}).Error
}

func(r *WishlistRepository) GetByUserID(userID uint) ([]domain.WishlistItem,error){
	var items []domain.WishlistItem
	err:=r.db.Where("user_id=?",userID).
    Preload("Product").
	Find(&items).Error

	return items,err
}



