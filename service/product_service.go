package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/repository"
	cloudinaryutil "backend/utils/utils/cloudinary"
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"strconv"
	"time"

	"gorm.io/gorm"
)


type ProductService struct {
	db          *gorm.DB
	productRepo repository.ProductRepository
	cloudinary  *cloudinaryutil.Client
	logger      *slog.Logger
}

func NewProductService(
	db *gorm.DB,
	productRepo repository.ProductRepository,
	cloudinary *cloudinaryutil.Client,
	logger *slog.Logger,
) *ProductService {
	if logger == nil {
		panic("ProductService requires a non-nil logger")
	}

	return &ProductService{
		db:          db,
		productRepo: productRepo,
		cloudinary:  cloudinary,
		logger:      logger,
	}
}

func (s *ProductService) CreateProduct(req dto.CreateProductRequest) (*domain.Product, error) {
	// --------------------
	// VALIDATIONS
	// --------------------
	if req.Name == "" || len(req.Name) < 2 || len(req.Name) > 100 {
		s.logger.Warn("CreateProduct failed: invalid name", "name", req.Name)
		return nil, ErrInvalidInput
	}

	if (req.Prices.H == nil || *req.Prices.H <= 0) && (req.Prices.F == nil || *req.Prices.F <= 0) {
		s.logger.Warn("CreateProduct failed: invalid prices", "prices", req.Prices)
		return nil, ErrInvalidInput
	}

	if req.Stock < 0 {
		s.logger.Warn("CreateProduct failed: invalid stock", "stock", req.Stock)
		return nil, ErrInvalidInput
	}

	if len(req.Description) > 500 {
		s.logger.Warn("CreateProduct failed: description too long")
		return nil, ErrInvalidInput
	}

	if req.DiscountPercent != nil {
		if *req.DiscountPercent <= 0 || *req.DiscountPercent >= 100 {
			s.logger.Warn("CreateProduct failed: invalid discount_percent", "value", *req.DiscountPercent)
			return nil, ErrInvalidInput
		}
	}

	if (req.OfferStart != nil || req.OfferEnd != nil) && req.DiscountPercent == nil {
		s.logger.Warn("CreateProduct failed: offer dates without discount")
		return nil, ErrInvalidInput
	}

	// --------------------
	// PARSE OFFER DATES
	// --------------------
	var offerStart, offerEnd *time.Time
	if req.OfferStart != nil {
		t, err := time.Parse(time.RFC3339, *req.OfferStart)
		if err != nil {
			s.logger.Warn("CreateProduct failed: invalid offer_start format", "value", *req.OfferStart, "error", err)
			return nil, ErrInvalidInput
		}
		offerStart = &t
	}
	if req.OfferEnd != nil {
		t, err := time.Parse(time.RFC3339, *req.OfferEnd)
		if err != nil {
			s.logger.Warn("CreateProduct failed: invalid offer_end format", "value", *req.OfferEnd, "error", err)
			return nil, ErrInvalidInput
		}
		offerEnd = &t
	}

	// --------------------
	// CREATE PRODUCT
	// --------------------
	product := &domain.Product{
		Name:            req.Name,
		Description:     req.Description,
		Stock:           req.Stock,
		CategoryID:      req.CategoryID,
		IsActive:        true,
		DiscountPercent: req.DiscountPercent,
		OfferStart:      offerStart,
		OfferEnd:        offerEnd,
	}

	tx := s.db.Begin()
	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		s.logger.Error("CreateProduct failed: db error", "name", req.Name, "err", err)
		return nil, err
	}

	// --------------------
	// CREATE PRICES (H/F)
	// --------------------
	prices := []domain.ProductPrice{}
	if req.Prices.H != nil {
		prices = append(prices, domain.ProductPrice{ProductID: product.ID, Type: "H", Price: *req.Prices.H})
	}
	if req.Prices.F != nil {
		prices = append(prices, domain.ProductPrice{ProductID: product.ID, Type: "F", Price: *req.Prices.F})
	}
	if len(prices) > 0 {
		if err := tx.Create(&prices).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		product.Prices = prices

	// Compute FinalPrice for the product (e.g., use "H" type as default)
	product.FinalPrice = product.CalculatePrice("H", time.Now())
	}

	// --------------------
	// CREATE IMAGES
	// --------------------
	images := []domain.ProductImage{}
	for i, url := range req.Images {
		images = append(images, domain.ProductImage{
			ProductID: product.ID,
			URL:       url,
			IsPrimary: i == 0, // first image primary
		})
	}
	if len(images) > 0 {
		if err := tx.Create(&images).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// --------------------
	// CREATE VARIANTS
	// --------------------
	variants := []domain.ProductVariant{}
	for _, v := range req.Variants {
		var variantPrices []domain.ProductPrice

		if v.H != nil {
			variantPrices = append(variantPrices, domain.ProductPrice{ProductID: product.ID, Type: "H", Price: *v.H})
		}
		if v.F != nil {
			variantPrices = append(variantPrices, domain.ProductPrice{ProductID: product.ID, Type: "F", Price: *v.F})
		}

		if len(variantPrices) > 0 {
			if err := tx.Create(&variantPrices).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		var priceID *uint
		if len(variantPrices) > 0 {
			priceID = &variantPrices[0].ID
		}

		variants = append(variants, domain.ProductVariant{
			ProductID: product.ID,
			Size:      v.Size,
			PriceID:   priceID,
			Stock:     v.Stock,
		})
	}

	if len(variants) > 0 {
		if err := tx.Create(&variants).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// --------------------
	// COMMIT TRANSACTION
	// --------------------
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.logger.Info("CreateProduct success", "product_id", product.ID)
	return product, nil
}


// func (s *ProductService) ListProducts() ([]*domain.Product, error) {
// 	products, err := s.productRepo.List()
// 	if err != nil {
// 		s.logger.Error("ListProducts failed", "err", err)
// 		return nil, err
// 	}
// 	return products, nil
// }


func (s *ProductService) ListProducts() ([]dto.ProductResponse, error) {
	products, err := s.productRepo.List()
	if err != nil {
		return nil, err
	}

	var responses []dto.ProductResponse

	now := time.Now()

	for _, p := range products {
		p.FinalPrice = p.CalculatePrice("H", now)
		responses = append(responses, mapProductToResponse(*p))
	}

	return responses, nil
}



// func (s *ProductService) GetProduct(id uint) (*domain.Product, error) {
// 	if id == 0 {
// 		s.logger.Warn("GetProduct failed: invalid productID", "product_id", id)
// 		return nil, ErrInvalidInput
// 	}

// 	product, err := s.productRepo.GetByID(id)
// 	if err != nil || product == nil {
// 		s.logger.Warn(
// 			"GetProduct failed: product not found",
// 			"product_id", id,
// 			"err", err,
// 		)
// 		return nil, ErrNotFound
// 	}
//     	now := time.Now()
// 	product.FinalPrice = product.CalculatePrice("H",now)

// 	return product, nil
// }

func (s *ProductService) GetProduct(id uint) (*dto.ProductResponse, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	product, err := s.productRepo.GetByID(id)
	if err != nil || product == nil {
		return nil, ErrNotFound
	}

	now := time.Now()
	product.FinalPrice = product.CalculatePrice("H", now)

	response := mapProductToResponse(*product)

	return &response, nil
}



func (s *ProductService) UpdateProduct(productID uint, req dto.UpdateProductRequest) error {
	if productID == 0 {
		return ErrInvalidInput
	}

	product, err := s.productRepo.GetByID(productID)
	if err != nil || product == nil {
		return ErrNotFound
	}

	// ---------------- BASIC FIELDS ----------------
	if req.Name != nil {
		if len(*req.Name) < 2 || len(*req.Name) > 100 {
			return ErrInvalidInput
		}
		product.Name = *req.Name
	}

	if req.Description != nil {
		if len(*req.Description) > 500 {
			return ErrInvalidInput
		}
		product.Description = *req.Description
	}

	if req.Stock != nil {
		if *req.Stock < 0 {
			return ErrInvalidInput
		}
		product.Stock = *req.Stock
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	// ---------------- PRICE (maps to H) ----------------
	if req.Price != nil {
		if *req.Price <= 0 || *req.Price > 1_000_000 {
			return ErrInvalidInput
		}

		found := false
		for i := range product.Prices {
			if product.Prices[i].Type == "H" {
				product.Prices[i].Price = *req.Price
				found = true
				break
			}
		}

		if !found {
			product.Prices = append(product.Prices, domain.ProductPrice{
				ProductID: product.ID,
				Type:      "H",
				Price:     *req.Price,
			})
		}
	}

	// ---------------- DISCOUNT ----------------
	layout := "2006-01-02T15:04:05Z07:00"

	if req.DiscountPercent != nil {
		if *req.DiscountPercent <= 0 || *req.DiscountPercent >= 100 {
			return ErrInvalidInput
		}
		product.DiscountPercent = req.DiscountPercent
	}

	if req.OfferStart != nil {
		start, err := time.Parse(layout, *req.OfferStart)
		if err != nil {
			return ErrInvalidInput
		}
		product.OfferStart = &start
	}

	if req.OfferEnd != nil {
		end, err := time.Parse(layout, *req.OfferEnd)
		if err != nil {
			return ErrInvalidInput
		}
		product.OfferEnd = &end
	}

	// ---------------- SAVE (includes prices) ----------------
	if err := s.productRepo.Update(product); err != nil {
	s.logger.Error("UpdateProduct failed", "product_id", productID, "err", err)
	return err
}

	s.logger.Info("UpdateProduct success", "product_id", productID)
	return nil
}




func (s *ProductService) DeleteProduct(productID uint) error {
	if productID == 0 {
		s.logger.Warn("DeleteProduct failed: invalid productID", "product_id", productID)
		return ErrInvalidInput
	}

	product, err := s.productRepo.GetByID(productID)
	if err != nil || product == nil {
		return ErrNotFound
	}

	if err := s.productRepo.Delete(productID); err != nil {
		s.logger.Error(
			"DeleteProduct failed: db error",
			"product_id", productID,
			"err", err,
		)
		return err
	}

	s.logger.Info("DeleteProduct success", "product_id", productID)
	return nil
}

func(s *ProductService)ListActive(r dto.ProductListQuery)([]domain.Product,error){
	allowedSort:= map[string]bool{
		"price": true,
		"name": true,
		"created_at": true,
	}
	if !allowedSort[r.Sort]{
		r.Sort="created_at"
	}
	if r.Page <= 0 {
	r.Page = 1
    }

	if r.Limit<=0{
		r.Limit=10
	}
	if r.Limit>50{
		r.Limit=50
	}
	if r.Order!="asc" && r.Order!="desc"{
		r.Order="desc"
	}
    if !r.ShowInactive {
		isActive := true
		r.IsActive = &isActive
	}
    
   products, err := s.productRepo.ListFiltered(r)
	if err != nil {
		s.logger.Error("ListFiltered failed", "err", err)
		return nil, err
	}
	
	// Calculate dynamic prices for each product
	now := time.Now()
	for i := range products {
		products[i].FinalPrice = products[i].CalculatePrice("H",now)
	}
	
    return products, nil
}

//----------------------------------------------

type CategoryService struct{
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService{
	return &CategoryService{repo: repo}
}

func(c *CategoryService)Create(name string)(*domain.Category,error){
   category:= &domain.Category{Name: name}
   return category,c.repo.Create(category)
}

func(c *CategoryService)List()([]domain.Category,error){
	return  c.repo.List()
}


func (s *ProductService) UploadProductImage(
	ctx context.Context,
	productID uint,
	fileHeader *multipart.FileHeader,
	isPrimary bool,
) (*domain.ProductImage, error) {

	file, err := fileHeader.Open()
	if err != nil {
		s.logger.Error("UploadProductImage: failed to open file", "err", err)
		return nil, err
	}
	defer file.Close()

	folder := "products/" + strconv.Itoa(int(productID))

	imageURL, publicID, err := s.cloudinary.UploadImage(ctx, file, folder, productID)
	if err != nil {
		s.logger.Error("UploadProductImage: cloudinary upload failed", "err", err)
		return nil, err
	}

	if imageURL == "" || publicID == "" {
		s.logger.Error("Cloudinary upload returned empty URL or publicID")
		return nil, errors.New("Cloudinary upload failed")
	}

	// Save to product_images table
	productImage := &domain.ProductImage{
		ProductID: productID,
		URL:       imageURL,
		PublicID:  publicID,
		IsPrimary: isPrimary,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to DB
	err = s.productRepo.AddImage(productID, imageURL, publicID, isPrimary)
	if err != nil {
		s.logger.Error("UploadProductImage: failed to save image to DB", "err", err)
		return nil, err
	}

	s.logger.Info("UploadProductImage: success", "productID", productID, "url", imageURL, "publicID", publicID)
	return productImage, nil
}


func (s *ProductService) DeleteProductImage(
	ctx context.Context,
	imageID uint,
) error {

	if imageID == 0 {
		return ErrInvalidInput
	}

	// Get image from DB first
	image, err := s.productRepo.GetImageByID(imageID)
	if err != nil || image == nil {
		return ErrNotFound
	}

	// 1️⃣ Delete from Cloudinary
	err = s.cloudinary.DeleteImage(ctx, image.PublicID)
	if err != nil {
		s.logger.Error("Cloudinary delete failed", "err", err)
		return err
	}

	// 2️⃣ Delete from DB
	err = s.productRepo.DeleteImage(imageID)
	if err != nil {
		s.logger.Error("DB delete image failed", "err", err)
		return err
	}

	s.logger.Info("Image deleted successfully", "imageID", imageID)

	return nil
}
