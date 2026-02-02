package repository

import "backend/internal/domain"

type ProductReader interface {
	GetByID(id uint) (*domain.Product, error)
}
