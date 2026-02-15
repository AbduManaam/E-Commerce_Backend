package repository

import (
	"backend/handler/dto"
)

type heroRepo struct{}
func NewHeroRepo() HeroRepository { return &heroRepo{} }
func (r *heroRepo) GetHero() (*dto.HeroBanner, error) {
	return &dto.HeroBanner{
		ID: "hero1", Title: "Delicious Meals for Every Craving", Subtitle: "Enjoy 25% Off Today!",
		Price: 4.99, Discount: 25, ImageURL: "/images/bg.png", CTA: "Shop Now",
	}, nil
}


type featureRepo struct{}
func NewFeatureRepo() FeatureRepository { return &featureRepo{} }

func (r *featureRepo) GetAllFeatures() ([]*dto.Feature, error) {
	return []*dto.Feature{
		{ID: "f1", Title: "Fast Delivery", Desc: "Get your food in record time", Icon: "FaTruck"},
		{ID: "f2", Title: "Secure Payments", Desc: "100% secure payment options", Icon: "FaLock"},
		{ID: "f3", Title: "24/7 Support", Desc: "We are here to help anytime", Icon: "FaHeadset"},
	}, nil
}

type reviewRepo struct{}
func NewReviewRepo() ReviewRepository { return &reviewRepo{} }

func (r *reviewRepo) GetReviews() ([]*dto.Review, error) {
	return []*dto.Review{
		{ID: "r1", Name: "Donald Jackman", Role: "Content Creator", Rating: 5, Comment: "Incredibly user-friendly.", AvatarURL: "https://images.unsplash.com/photo-1633332755192-727a05c4013d?q=80&w=100"},
		{ID: "r2", Name: "Richard Nelson", Role: "Instagram Influencer", Rating: 5, Comment: "Amazing app for content creation.", AvatarURL: "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?q=80&w=100"},
		{ID: "r3", Name: "James Washington", Role: "Digital Content Creator", Rating: 5, Comment: "Highly recommend!", AvatarURL: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?q=80&w=100"},
	}, nil
}
