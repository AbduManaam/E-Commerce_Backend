package dto

type CreateAddressRequest struct {
	FullName  string `json:"full_name" validate:"required"`
	Phone     string `json:"phone" validate:"required"`
	Address   string `json:"address" validate:"required"`
	City      string `json:"city" validate:"required"`
	State     string `json:"state" validate:"required"`
	Country   string `json:"country"`
	ZipCode   string `json:"zip_code" validate:"required"`
	Landmark  string `json:"landmark"`
	IsDefault bool   `json:"is_default"`
}
