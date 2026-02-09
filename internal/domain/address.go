package domain

type Address struct {
    ID       uint   `gorm:"primaryKey"`
    UserID   uint   `gorm:"not null;index"`
    FullName string `gorm:"not null"`
    Phone    string `gorm:"not null"`
    Address  string `gorm:"type:text;not null"`
    City     string `gorm:"not null"`
    State    string `gorm:"not null"`
    Country  string `gorm:"not null;default:'India'"`
    ZipCode  string `gorm:"not null"`
    Landmark string
    IsDefault bool   `gorm:"default:false"`
    
}