package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type Tx = *gorm.DB

type ProductReader interface {
	GetByID(id uint) (*domain.Product, error)
	GetByIDForUpdate(tx *gorm.DB, id uint) (*domain.Product, error)
}

type ProductWriter interface {
	Update(p *domain.Product) error
	UpdateTx(tx *gorm.DB, product *domain.Product) error
}