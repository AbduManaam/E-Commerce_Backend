package dto

type ProductListQuery struct{
	Category string
	Sort     string
	Order    string
	Page     int
	Limit    int
	MinPrice *float64
    MaxPrice *float64
}