package domain

import "time"

type User struct {
	ID        uint
	Name      string
	Email     string
	Password  string
	Role      string	
	IsVerified bool
	IsBlocked  bool
	OTP       string    `json:"-" gorm:"size:255"`
	OTPExpiry time.Time `json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}