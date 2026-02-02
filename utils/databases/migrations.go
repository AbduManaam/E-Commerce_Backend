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
		&domain.Order{},
		&domain.CartItem{},
		&domain.WishlistItem{},
	)

	if err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	log.Println("✅ Database migration completed")
}