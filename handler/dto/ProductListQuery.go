package dto

type ProductListQuery struct{
	CategoryID *uint    `json:"category_id"`
	Sort     string      `json:"sort"` 
	Order    string     `json:"order"`
	Search   string    `json:"search"`
	Page     int       `json:"page"`
	Limit    int        `json:"limit"`
	MinPrice *float64   `json:"min_price"`
    MaxPrice *float64   `json:"max_price"`
	OnlyActiveOffers bool     `json:"only_active_offers"`
    ShowInactive    bool     `form:"show_inactive"`      // For admin
	IsActive        *bool    `form:"is_active"` 
}

