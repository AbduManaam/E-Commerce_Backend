package dto

type CreateAddressRequest struct {
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	ZipCode   string `json:"zip_code"`
	Landmark  string `json:"landmark"`
	IsDefault bool   `json:"is_default"`
}
