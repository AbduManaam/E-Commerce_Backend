package domain

import "time"

type Product struct {
    ID          uint
    Name        string
    Description string
    Price       float64
    Stock       int
    CategoryID  uint
    Category    Category `gorm:"foreignKey:CategoryID"`
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}


type Category struct {
    ID   uint
    Name string
}
