package dto

type ProductResponse struct {
    ID           uint                `json:"id"`
    Name         string              `json:"name"`
    Price        float64             `json:"price"`
    CategoryID   uint                `json:"category_id"`
    CategoryName string              `json:"category_name"`
    Images       []ProductImageDTO   `json:"images"`
}


type ProductImageDTO struct {
    URL string `json:"url"`
}
