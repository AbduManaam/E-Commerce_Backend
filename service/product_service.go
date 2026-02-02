package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/repository"
)

type ProductService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) CreateProduct(req dto.CreateProductRequest) (*domain.Product, error) {

	if req.Name == "" || len(req.Name) < 2 || len(req.Name) > 100 {
		return nil, ErrInvalidInput
	}
	if req.Price <= 0 || req.Price > 1_000_000 {
		return nil, ErrInvalidInput
	}
	if req.Stock < 0 {
		return nil, ErrInvalidInput
	}
	if len(req.Description) > 500 {
		return nil, ErrInvalidInput
	}
	
	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		IsActive:    true,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) ListProducts() ([]*domain.Product, error) {
	return s.productRepo.List()
}

func(s *ProductService) GetProduct(id uint)(*domain.Product,error){

	product,err:= s.productRepo.GetByID(id)
	if err!=nil{
		return nil,err
	}
	return product,nil
}
func (s *ProductService) UpdateProduct(
	productID uint,
	req dto.UpdateProductRequest,
) error {

	if productID == 0 {
		return ErrInvalidInput
	}

	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return err
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

	return s.productRepo.Update(product)
}



func (s *ProductService) DeleteProduct(productID uint) error {
	if productID == 0 {
		return ErrInvalidInput
	}

	return s.productRepo.Delete(productID)
}
