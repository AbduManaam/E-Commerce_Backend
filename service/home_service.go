package service

import (
	"backend/handler/dto"
	"backend/repository"
)

type HomeService struct {
	heroRepo    repository.HeroRepository
	productRepo repository.ProductRepository
	featureRepo repository.FeatureRepository
	reviewRepo  repository.ReviewRepository
}

func NewHomeService(
	heroRepo repository.HeroRepository,
	productRepo repository.ProductRepository,
	featureRepo repository.FeatureRepository,
	reviewRepo repository.ReviewRepository,
) *HomeService {
	return &HomeService{
		heroRepo:    heroRepo,
		productRepo: productRepo,
		featureRepo: featureRepo,
		reviewRepo:  reviewRepo,
	}
}

type HomeResponse struct {
	Hero        *dto.HeroBanner   `json:"hero"`
	NewArrivals []*dto.Product    `json:"new_arrivals"`
	Features    []*dto.Feature    `json:"features"`
	Reviews     []*dto.Review     `json:"reviews"`
}

func (s *HomeService) GetHomeData() (*HomeResponse, error) {

	hero, err := s.heroRepo.GetHero()
	if err != nil {
		return nil, err
	}

	products, err := s.productRepo.GetNewArrivals(10)
	if err != nil {
		return nil, err
	}

	features, err := s.featureRepo.GetAllFeatures()
	if err != nil {
		return nil, err
	}

	reviews, err := s.reviewRepo.GetReviews()
	if err != nil {
		return nil, err
	}

	return &HomeResponse{
		Hero:        hero,
		NewArrivals: products,
		Features:    features,
		Reviews:     reviews,
	}, nil
}
