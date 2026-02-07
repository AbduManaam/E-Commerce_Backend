package repository

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type productRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// Constructor
func NewProductRepository(
	db *gorm.DB,
	logger *slog.Logger,
) ProductRepositoryInterface {
	return &productRepository{
		db:     db,
		logger: logger,
	}
}

// -----------------------------------------------------

func (r *productRepository) Create(p *domain.Product) error {
	err := r.db.Create(p).Error
	if err != nil {
		r.logger.Error(
			"product create failed",
			"name", p.Name,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"product created",
		"product_id", p.ID,
		"name", p.Name,
	)
	return nil
}

func (r *productRepository) GetByID(id uint) (*domain.Product, error) {
	var product domain.Product

	err := r.db.
	Preload("Category").
	First(&product, id).Error
	if err != nil {
		r.logger.Error(
			"product get by id failed",
			"product_id", id,
			"err", err,
		)
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) List() ([]*domain.Product, error) {
	var products []*domain.Product

	err := r.db.
	Preload("Category").
	Find(&products).Error
	if err != nil {
		r.logger.Error(
			"product list failed",
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"product list fetched",
		"count", len(products),
	)
	return products, nil
}

func (r *productRepository) Update(p *domain.Product) error {
	err := r.db.Save(p).Error
	if err != nil {
		r.logger.Error(
			"product update failed",
			"product_id", p.ID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"product updated",
		"product_id", p.ID,
	)
	return nil
}

func (r *productRepository) Delete(id uint) error {
	err := r.db.Delete(&domain.Product{}, id).Error
	if err != nil {
		r.logger.Error(
			"product delete failed",
			"product_id", id,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"product deleted",
		"product_id", id,
	)
	return nil
}

func (r *productRepository) ListFiltered(q dto.ProductListQuery) ([]domain.Product, error) {
	var products []domain.Product

	db := r.db.Model(&domain.Product{}).
		Preload("Category").
		Where("is_active = ?", true)

	// Filter by category
	if q.CategoryID != nil {
		db = db.Where("category_id = ?", *q.CategoryID)
	}

	// Filter by price range
	if q.MinPrice != nil && *q.MinPrice > 0 {
		db = db.Where("price >= ?", *q.MinPrice)
	}
	if q.MaxPrice != nil && *q.MaxPrice > 0 {
		db = db.Where("price <= ?", *q.MaxPrice)
	}

	// Filter by search term
	if q.Search != "" {
		like := "%" + q.Search + "%"
		db = db.Where(
			"products.name ILIKE ? OR products.description ILIKE ?",
			like, like,
		)
	}

	// Filter by active offers
	if q.OnlyActiveOffers {
		now := time.Now()
		db = db.Where(
			"(discount_percent IS NOT NULL OR discount_amount IS NOT NULL) AND " +
				"(offer_start IS NULL OR offer_start <= ?) AND " +
				"(offer_end IS NULL OR offer_end >= ?)",
			now, now,
		)
	}

	// Pagination
	offset := (q.Page - 1) * q.Limit

	err := db.
		Order(q.Sort + " " + q.Order).
		Limit(q.Limit).
		Offset(offset).
		Find(&products).
		Error

	return products, err
}

//------------------------------------------------

type CategoryRepository struct{
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository{
   return &CategoryRepository{db:db}
}

func (c *CategoryRepository)Create(d *domain.Category)error{
	return c.db.Create(d).Error
}

func (c *CategoryRepository) List() ([]domain.Category,error){
   var category []domain.Category
   err:=c.db.Find(&category).Error
   return category,err
}