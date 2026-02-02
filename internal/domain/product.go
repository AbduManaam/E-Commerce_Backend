package domain

import "time"

type Product struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	Stock       int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
