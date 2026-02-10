package databases

import (
	"backend/internal/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
	&domain.User{},
	&domain.Product{},
	&domain.Category{},
	&domain.Address{},        // user addresses
	&domain.OrderAddress{},   // 🔥 REQUIRED
	&domain.Order{},
	&domain.OrderItem{},      // 🔥 REQUIRED
	&domain.CartItem{},
	&domain.WishlistItem{},
	&domain.Payment{},
	&domain.RefreshToken{},
)

// err := db.AutoMigrate(
// 		&domain.User{},
// 		&domain.Product{},
// 		&domain.Category{},
// 		&domain.Order{},
// 		&domain.RefreshToken{}, 
// 		&domain.CartItem{},
// 		&domain.WishlistItem{},
// 		&domain.Payment{},
// 		&domain.Address{},
// 	)

	if err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	log.Println("✅ Database migration completed")
}