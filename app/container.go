
package app

import (
	"backend/config"
	"backend/handler"
	"backend/repository"
	"backend/service"
	"backend/utils/databases"
	"backend/utils/email"
)

type Container struct {
	UserHandler    *handler.UserHandler
	AdminHandler   *handler.AdminUserHandler
	AuthHandler    *handler.AuthHandler
	ProductHandler *handler.ProductHandler
	OrderHandler   *handler.OrderHandler
    CartHandler     *handler.CartHandler
	WishlistHandler *handler.WishlistHandler

	DBCleanup func() error
}

func BuildContainer(cfg *config.AppConfig) (*Container, error) {
	db := databases.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	databases.AutoMigrate(db)
	
	// repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(db)
    wishlistRepo := repository.NewWishlistRepository(db)

	// Initialize email package with config
	email.Init(cfg.SMTP)
	emailSvc := email.NewSMTPService()

	authSvc := service.NewAuthService(
	userRepo,
	&cfg.JWT,
	emailSvc,
)
	// services 
	userSvc := service.NewUserService(userRepo)
	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(
		orderRepo,
		productRepo,
	)
	cartSvc := service.NewCartService(
	cartRepo,
	productRepo,
    )

    wishlistSvc := service.NewWishlistService(
	wishlistRepo,
	productRepo,
    )
	cartHandler := handler.NewCartHandler(cartSvc)
    wishlistHandler := handler.NewWishlistHandler(wishlistSvc)

	// handlers
	return &Container{
		AuthHandler:    handler.NewAuthHandler(authSvc),
		AdminHandler:   handler.NewAdminUserHandler(userSvc),
		UserHandler:    handler.NewUserHandler(userSvc),
		ProductHandler: handler.NewProductHandler(productSvc),
		OrderHandler:   handler.NewOrderHandler(orderSvc),
		CartHandler:     cartHandler,
	    WishlistHandler: wishlistHandler,
		DBCleanup:      sqlDB.Close,
	}, nil
}