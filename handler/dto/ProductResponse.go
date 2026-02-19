package dto

import "backend/internal/domain"   // ← adjust import path to your project

type ProductResponse struct {
    ID           uint              `json:"id"`
    Name         string            `json:"name"`
    Price        float64           `json:"price"`
    CategoryID   uint              `json:"category_id"`
    CategoryName string            `json:"category_name"`
    Images       []ProductImageDTO `json:"images"`
}

type ProductImageDTO struct {
    URL string `json:"url"`
}

func ToProductResponse(p domain.Product) ProductResponse {
    var images []ProductImageDTO

    for _, img := range p.Images {
        images = append(images, ProductImageDTO{
            URL: img.URL,
        })
    }

    return ProductResponse{
        ID:           p.ID,
        Name:         p.Name,
        Price:        p.FinalPrice,
        CategoryID:   p.CategoryID,
        CategoryName: p.Category.Name, // make sure Category is preloaded
        Images:       images,
    }
}
