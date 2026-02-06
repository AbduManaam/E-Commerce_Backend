package dto

type ProductListQuery struct{
	CategoryID *uint
	Sort     string
	Order    string
	Page     int
	Limit    int
	MinPrice *float64
    MaxPrice *float64
}