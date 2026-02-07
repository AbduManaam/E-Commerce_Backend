package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/repository"
	"log/slog"

)


type ProductService struct {
	productRepo repository.ProductRepository
	logger      *slog.Logger
}

func NewProductService(
	productRepo repository.ProductRepository,
	logger *slog.Logger,
) *ProductService {
	if logger == nil {
		panic("ProductService requires a non-nil logger")
	}

	return &ProductService{
		productRepo: productRepo,
		logger:      logger,
	}
}


func (s *ProductService) CreateProduct(req dto.CreateProductRequest) (*domain.Product, error) {
	if req.Name == "" || len(req.Name) < 2 || len(req.Name) > 100 {
		s.logger.Warn("CreateProduct failed: invalid name", "name", req.Name)
		return nil, ErrInvalidInput
	}

	if req.Price <= 0 || req.Price > 1_000_000 {
		s.logger.Warn("CreateProduct failed: invalid price", "price", req.Price)
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

	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		IsActive:    true,
	}

	if err := s.productRepo.Create(product); err != nil {
		s.logger.Error(
			"CreateProduct failed: db error",
			"name", req.Name,
			"err", err,
		)
		return nil, err
	}

	s.logger.Info("CreateProduct success", "product_id", product.ID)
	return product, nil
}


func (s *ProductService) ListProducts() ([]*domain.Product, error) {
	products, err := s.productRepo.List()
	if err != nil {
		s.logger.Error("ListProducts failed", "err", err)
		return nil, err
	}
	return products, nil
}

func (s *ProductService) GetProduct(id uint) (*domain.Product, error) {
	if id == 0 {
		s.logger.Warn("GetProduct failed: invalid productID", "product_id", id)
		return nil, ErrInvalidInput
	}

	product, err := s.productRepo.GetByID(id)
	if err != nil || product == nil {
		s.logger.Warn(
			"GetProduct failed: product not found",
			"product_id", id,
			"err", err,
		)
		return nil, ErrNotFound
	}

	return product, nil
}


func (s *ProductService) UpdateProduct(productID uint, req dto.UpdateProductRequest) error {
	if productID == 0 {
		s.logger.Warn("UpdateProduct failed: invalid productID", "product_id", productID)
		return ErrInvalidInput
	}

	product, err := s.productRepo.GetByID(productID)
	if err != nil || product == nil {
		s.logger.Warn(
			"UpdateProduct failed: product not found",
			"product_id", productID,
		)
		return ErrNotFound
	}

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

	if req.Price != nil {
		if *req.Price <= 0 || *req.Price > 1_000_000 {
			return ErrInvalidInput
		}
		product.Price = *req.Price
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

	if err := s.productRepo.Update(product); err != nil {
		s.logger.Error(
			"UpdateProduct failed: db error",
			"product_id", productID,
			"err", err,
		)
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
	return s.productRepo.ListFiltered(r)
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