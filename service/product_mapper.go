package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
)

func mapProductToResponse(p domain.Product) dto.ProductResponse {
    images := make([]dto.ProductImageDTO, 0) // ← always returns []
    for _, img := range p.Images {
        images = append(images, dto.ProductImageDTO{
            URL: img.URL,
        })
    }

    return dto.ProductResponse{
        ID:           p.ID,
        Name:         p.Name,
        Price:        p.FinalPrice,
        CategoryID:   p.CategoryID,
        CategoryName: p.Category.Name, // VERY IMPORTANT
        Images:       images,
        Stock:        p.Stock,
		Prices:       mapPrices(p.Prices),
	}
}

func mapPrices(prices []domain.ProductPrice) []dto.PriceResponseDTO {
    result := make([]dto.PriceResponseDTO, 0)
    for _, p := range prices {
        result = append(result, dto.PriceResponseDTO{
            Size:  p.Type, // ← "H" or "F"
            Price: p.Price,
        })
    }
    return result
}