package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
)

func mapProductToResponse(p domain.Product) dto.ProductResponse {
    images := []dto.ProductImageDTO{}
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
    }
}
