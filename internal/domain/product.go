package domain

import "time"

type Product struct {
    ID          uint    `gorm:"primaryKey"`
    Name        string   `gorm:"not null"`
    Description string   
    Price       float64  `gorm:"not null"`
    Stock       int      `gorm:"not null"`
    CategoryID  uint     `gorm:"not null"`
    Category    Category `gorm:"foreignKey:CategoryID"`
    IsActive    bool     `gorm:"default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}


type Category struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"unique;not null"`
}
