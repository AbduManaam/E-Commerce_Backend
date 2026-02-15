package dto


type HeroBanner struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle string  `json:"subtitle"`
	Price    float64 `json:"price"`
	Discount int     `json:"discount"`
	ImageURL string  `json:"image_url"`
	CTA      string  `json:"cta"`
}

type Product struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Price    float64  `json:"price"`
	InStock  bool     `json:"in_stock"`
	Sizes    []string `json:"sizes"`
	ImageURL string   `json:"image_url"`
	Category string   `json:"category"`
}

type Feature struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Icon  string `json:"icon"`
}

type Review struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	AvatarURL string `json:"avatar_url"`
}
