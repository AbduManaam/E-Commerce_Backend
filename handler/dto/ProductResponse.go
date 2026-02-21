package dto

import "backend/internal/domain"   // ← adjust import path to your project

type ProductResponse struct {
    ID           uint              `json:"id"`
    Name         string            `json:"name"`
    Price        float64           `json:"price"`
    CategoryID   uint              `json:"category_id"`
    CategoryName string            `json:"category_name"`
    Images       []ProductImageDTO `json:"images"`
	Stock        int               `json:"stock"`
    Prices       []PriceResponseDTO`json:"prices"` // ← add type

}

type PriceResponseDTO struct {
    Size  string  `json:"size"`
    Price float64 `json:"price"`
}

type ProductImageDTO struct {
    URL string `json:"url"`
}
func ToProductResponse(p domain.Product) ProductResponse {
    var images []ProductImageDTO
    for _, img := range p.Images {
        images = append(images, ProductImageDTO{URL: img.URL})
    }

    prices := make([]PriceResponseDTO, 0)
    for _, pr := range p.Prices {
        prices = append(prices, PriceResponseDTO{
            Size:  pr.Type,
            Price: pr.Price,
        })
    }

    return ProductResponse{
        ID:           p.ID,
        Name:         p.Name,
        Price:        p.FinalPrice,
        CategoryID:   p.CategoryID,
        CategoryName: p.Category.Name,
        Images:       images,
        Stock:        p.Stock,
        Prices:       prices,
    }
}