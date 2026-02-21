package repository

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"fmt"
	"log/slog"
	"strings"
	"time"

	// "time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// Constructor
func NewProductRepository(
	db *gorm.DB,
	logger *slog.Logger,
) ProductRepository {
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
	Preload("Prices").
	Preload("Images").
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
		Preload("Prices").
		Preload("Images").
		Find(&products).Error
	if err != nil {
		r.logger.Error("product list failed", "err", err)
		return nil, err
	}

	r.logger.Info("product list fetched", "count", len(products))
	return products, nil
}


func (r *productRepository) Update(p *domain.Product) error {
	err := r.db.Session(&gorm.Session{FullSaveAssociations: true}).
	Save(p).Error
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


func (r *productRepository) ListFiltered(query dto.ProductListQuery) ([]domain.Product, int64, error) {
    var products []domain.Product
    var total int64
    
    // Start building the query
    db := r.db.Model(&domain.Product{})
    
    // ==========================================
    // APPLY FILTERS
    // ==========================================
    
    // Category filter
    if query.CategoryID != nil {
        db = db.Where("category_id = ?", *query.CategoryID)
    }
    
    // Search filter
    if query.Search != "" {
    searchPattern := "%" + strings.ToLower(query.Search) + "%"
    db = db.Joins("LEFT JOIN categories ON categories.id = products.category_id").
        Where("LOWER(products.name) LIKE ? OR LOWER(categories.name) LIKE ?",
            searchPattern, searchPattern)
}
    
    // Active/Inactive filter
    if query.IsActive != nil {
        db = db.Where("is_active = ?", *query.IsActive)
    }
    
    // Price range filters
    if query.MinPrice != nil {
        // You'll need to join with product_prices table
        db = db.Joins("LEFT JOIN product_prices ON products.id = product_prices.product_id")
        db = db.Where("product_prices.price >= ?", *query.MinPrice)
    }
    
    if query.MaxPrice != nil {
        if query.MinPrice == nil {
            db = db.Joins("LEFT JOIN product_prices ON products.id = product_prices.product_id")
        }
        db = db.Where("product_prices.price <= ?", *query.MaxPrice)
    }
    
    // Offer filters
    if query.OnlyActiveOffers {
        now := time.Now()
        db = db.Where("discount_percent > 0")
        db = db.Where("(offer_start IS NULL OR offer_start <= ?)", now)
        db = db.Where("(offer_end IS NULL OR offer_end >= ?)", now)
    }
    
    // ==========================================
    // ✅ COUNT TOTAL BEFORE PAGINATION
    // ==========================================
    countDB := db.Session(&gorm.Session{})
    if err := countDB.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // ==========================================
    // APPLY PAGINATION
    // ==========================================
    offset := (query.Page - 1) * query.Limit
    db = db.Offset(offset).Limit(query.Limit)
    
    // ==========================================
    // APPLY SORTING
    // ==========================================
    sortColumn := query.Sort
    sortOrder := query.Order
    
    // Handle special case for price sorting
    if sortColumn == "price" {
        // Sort by the H type price
        db = db.Joins("LEFT JOIN product_prices AS pp ON products.id = pp.product_id AND pp.type = 'H'")
        db = db.Order("pp.price " + sortOrder)
    } else {
        db = db.Order(sortColumn + " " + sortOrder)
    }
    
    // ==========================================
    // PRELOAD RELATIONSHIPS
    // ==========================================
    db = db.Preload("Images").Preload("Prices").Preload("Category").Preload("Variants")
    
    // ==========================================
    // EXECUTE QUERY
    // ==========================================
    if err := db.Find(&products).Error; err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
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

func (r *productRepository) GetByIDForUpdate(
	tx *gorm.DB,
	id uint,
) (*domain.Product, error) {

	var product domain.Product

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Prices").        // ✅ ADD THIS - load prices
		Preload("Images").        // ✅ ADD THIS - load images
		Preload("Category").      
		First(&product, id).
		Error

	if err != nil {
		r.logger.Error(
			"product get for update failed",
			"product_id", id,
			"err", err,
		)
		return nil, err
	}

	return &product, nil
}


func (r *productRepository) UpdateTx(
	tx *gorm.DB,
	product *domain.Product,
) error {

	if err := tx.Save(product).Error; err != nil {
		r.logger.Error(
			"product update tx failed",
			"product_id", product.ID,
			"err", err,
		)
		return err
	}

	return nil
}





func (r *productRepository) GetNewArrivals(limit int) ([]*dto.Product, error) {
	var products []domain.Product
	
	// ✅ Fetch real products from database
	err := r.db.
		Preload("Images").           // Load product images
		Preload("Prices").           // Load product prices
		Preload("Category").         // Load category
		Where("is_active = ?", true). // Only active products
		Order("created_at DESC").     // Newest first
		Limit(limit).
		Find(&products).Error
	
	if err != nil {
		r.logger.Error("GetNewArrivals failed", "err", err)
		return nil, err
	}
	
	// Convert domain.Product to dto.Product
	var dtoProducts []*dto.Product
	
	for _, p := range products {
		// Calculate final price
		p.FinalPrice = p.CalculatePrice("H", time.Now())
		
		// Build DTO
		dtoProduct := &dto.Product{
			ID:           fmt.Sprintf("%d", p.ID), // Convert uint to string
			Name:         p.Name,
			Price:        p.FinalPrice,
			InStock:      p.Stock > 0,
			Category:     p.Category.Name,
		}
		
		// Add images
		if len(p.Images) > 0 {
			dtoProduct.ImageURL = p.Images[0].URL // First image as main
			// If your dto.Product has an Images array field, populate it too
		}
		
		// Add sizes (if you have variants)
		if len(p.Variants) > 0 {
			sizes := make([]string, len(p.Variants))
			for i, v := range p.Variants {
				sizes[i] = v.Size
			}
			dtoProduct.Sizes = sizes
		} else {
			dtoProduct.Sizes = []string{"default"}
		}
		
		dtoProducts = append(dtoProducts, dtoProduct)
	}
	
	r.logger.Info("GetNewArrivals success", "count", len(dtoProducts))
	return dtoProducts, nil
}



// Get products filtered by optional min/max price
func (r *productRepository) GetProductsByPrice(minPrice, maxPrice *float64) ([]domain.Product, error) {
    query := r.db.Model(&domain.Product{}).
        Joins("JOIN product_prices pp ON pp.product_id = products.id")

    if minPrice != nil {
        query = query.Where("pp.price >= ?", *minPrice)
    }
    if maxPrice != nil {
        query = query.Where("pp.price <= ?", *maxPrice)
    }

    var products []domain.Product
    if err := query.Distinct("products.id").Find(&products).Error; err != nil {
        return nil, err
    }

    return products, nil
}

// Cloudinary

// func (r *productRepository) AddImage(
// 	productID uint,
// 	url string,
// 	publicID string,
// 	isPrimary bool,
// ) error {

// 	query := `
// 	INSERT INTO product_images (product_id, image_url, public_id, is_primary)
// 	VALUES ($1,$2,$3,$4)
// 	`

// 	return  r.db.Exec(query, productID, url, publicID, isPrimary).Error
	
// }

func (r *productRepository) AddImage(productID uint, url string, publicID string, isPrimary bool) error {
    image := domain.ProductImage{
        ProductID: productID,
        URL:       url,      // GORM maps this to image_url column ✅
        PublicID:  publicID,
        IsPrimary: isPrimary,
    }
    return r.db.Create(&image).Error  // ← use GORM not raw SQL
}


func (r *productRepository) GetImageByID(id uint) (*domain.ProductImage, error) {
	var image domain.ProductImage
	err := r.db.First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *productRepository) DeleteImage(id uint) error {
	return r.db.Delete(&domain.ProductImage{}, id).Error
}
