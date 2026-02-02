package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)




type productRepository struct {
	db *gorm.DB
}


// func NewProductRepository(db *gorm.DB) ProductRepositoryInterface  {
// 	var repo ProductRepository = &productRepository{db: db}
// 	return repo
// }

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &productRepository{db: db} 
}


func (r *productRepository) Create(p *domain.Product) error {
	return r.db.Create(p).Error
}

func (r *productRepository) GetByID(id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error
	return &product, err
}

func (r *productRepository) List() ([]*domain.Product, error) {
	var products []*domain.Product
	err := r.db.Find(&products).Error
	return products, err
}

func (r *productRepository) Update(p *domain.Product) error {
	return r.db.Save(p).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Product{}, id).Error
}
