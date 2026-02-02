package dto

type AdminUpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
	Role string `json:"role" validate:"required,oneof=USER ADMIN"`
}


type UpdateProductRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0,lt=1000000"`
	Stock       *int     `json:"stock" validate:"omitempty,gte=0"`
	IsActive    *bool    `json:"is_active"`
}

