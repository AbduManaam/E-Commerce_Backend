package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}



func NewCartRepository(db *gorm.DB) *CartRepository{
	return &CartRepository{db:db}
}

func (r *CartRepository) GetorCreateCart(userID uint) (*domain.Cart,error){
   
	var cart domain.Cart
	err:= r.db.Where("user_id=?",userID).
	Preload("Items").
	First(&cart).Error

	if err== gorm.ErrRecordNotFound{
		cart= domain.Cart{UserID: userID}
		return &cart,r.db.Create(&cart).Error
	}
	return &cart,err
}

func(r *CartRepository) FindItem(CartID,ItemID uint) (*domain.CartItem,error)  {
	
	var cartItem domain.CartItem
	err:= r.db.Where("cart_id = ? AND id = ?",CartID,ItemID).
	First(&cartItem).Error
	return &cartItem,err
}

func(r *CartRepository) Save(item *domain.CartItem)error{
	return r.db.Save(item).Error
}
func(r *CartRepository) Delete(item *domain.CartItem)error{
	return r.db.Delete(item).Error
}
