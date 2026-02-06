
package app

import (
	"backend/config"
	"backend/handler"
	"backend/repository"
	"backend/service"
	"backend/utils/databases"
	"backend/utils/email"
	"backend/utils/logging"
	"log"
	"os"
)

type Container struct {
	UserHandler     *handler.UserHandler
	AdminHandler    *handler.AdminUserHandler
	AuthHandler     *handler.AuthHandler
	ProductHandler  *handler.ProductHandler
	OrderHandler    *handler.OrderHandler
	CartHandler     *handler.CartHandler
	WishlistHandler *handler.WishlistHandler
    CategoryHandler *handler.CategoryHandler


	DBCleanup func() error
}

func BuildContainer(cfg *config.AppConfig) (*Container, error) {

	logging.Init(cfg.Environment)
	logger := logging.Logger

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
	repoLogger := log.New(os.Stdout, "[SERVICE] ", log.LstdFlags|log.Lshortfile)


	authRepo := repository.NewAuthRepository(db)
	userRepo := repository.NewUserRepository(db, logger)
	productRepo := repository.NewProductRepository(db, logger)
	orderRepo := repository.NewOrderRepository(db, logger)
	cartRepo := repository.NewCartRepository(db, logger)
	wishlistRepo := repository.NewWishlistRepository(db, logger)

	categoryRepo:= repository.NewCategoryRepository(db)
	categorySvc := service.NewCategoryService(categoryRepo)
    categoryHandler := handler.NewCategoryHandler(categorySvc)



	// Email
	email.Init(cfg.SMTP)
	emailSvc := email.NewSMTPService()

	// Services
	authSvc := service.NewAuthService(
		userRepo,
		authRepo,
		&cfg.JWT,
		emailSvc,
	)

	userSvc := service.NewUserService(userRepo,logger)
	productSvc := service.NewProductService(productRepo,logger)

	orderSvc := service.NewOrderService(
		orderRepo,
		productRepo,
		repoLogger,
	)

	cartSvc := service.NewCartService(
		cartRepo,
		productRepo,
		repoLogger,
	)

	wishlistSvc := service.NewWishlistService(
		wishlistRepo,
		productRepo,
		repoLogger,
	)

	// Handlers
	return &Container{
		AuthHandler:     handler.NewAuthHandler(authSvc),
		AdminHandler:    handler.NewAdminUserHandler(userSvc),
		UserHandler:     handler.NewUserHandler(userSvc),
		ProductHandler:  handler.NewProductHandler(productSvc),
		OrderHandler:    handler.NewOrderHandler(orderSvc),
		CartHandler:     handler.NewCartHandler(cartSvc),
		WishlistHandler: handler.NewWishlistHandler(wishlistSvc),
        CategoryHandler: categoryHandler, 
		DBCleanup:       sqlDB.Close,
	}, nil
}
