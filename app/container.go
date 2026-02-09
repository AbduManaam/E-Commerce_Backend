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
	// Repositories
	UserRepo repository.UserRepository

	// Handlers
	AuthHandler     *handler.AuthHandler
	UserHandler     *handler.UserHandler
	AdminHandler    *handler.AdminUserHandler
	ProductHandler  *handler.ProductHandler
	OrderHandler    *handler.OrderHandler
	CartHandler     *handler.CartHandler
	WishlistHandler *handler.WishlistHandler
	CategoryHandler *handler.CategoryHandler
	AddressHandler  *handler.AddressHandler // ✅ ADDED
	PaymentHandler  *handler.PaymentHandler

	DBCleanup func() error
}

func BuildContainer(cfg *config.AppConfig) (*Container, error) {
	// ------------------------------------------------
	// Logger
	// ------------------------------------------------
	logging.Init(cfg.Environment)
	logger := logging.Logger

	// ------------------------------------------------
	// Database
	// ------------------------------------------------
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

	// ------------------------------------------------
	// Repositories
	// ------------------------------------------------
	authRepo := repository.NewAuthRepository(db)
	userRepo := repository.NewUserRepository(db, logger)
	productRepo := repository.NewProductRepository(db, logger)
	orderRepo := repository.NewOrderRepository(db, logger)
	wishlistRepo := repository.NewWishlistRepository(db, logger)
	categoryRepo := repository.NewCategoryRepository(db)
	addressRepo := repository.NewAddressRepository(db) // ✅ ADDED
	paymentRepo := repository.NewPaymentRepository(db)

	var cartRepo repository.CartRepositoryInterface =
		repository.NewCartRepository(db, logger)

	var productReader repository.ProductReader = productRepo
	var productWriter repository.ProductWriter = productRepo

	// ------------------------------------------------
	// Email
	// ------------------------------------------------
	email.Init(cfg.SMTP)
	emailSvc := email.NewSMTPService()

	// ------------------------------------------------
	// Services
	// ------------------------------------------------
	authSvc := service.NewAuthService(
		userRepo,
		authRepo,
		&cfg.JWT,
		emailSvc,
	)

	userSvc := service.NewUserService(userRepo, logger)
	productSvc := service.NewProductService(productRepo, logger)

	orderSvc := service.NewOrderService(
		orderRepo,
		productReader,
		productWriter,
		cartRepo,
		repoLogger,
	)

	cartSvc := service.NewCartService(
		cartRepo,
		productReader,
		productWriter,
		repoLogger,
	)

	wishlistSvc := service.NewWishlistService(
		wishlistRepo,
		productReader,
		repoLogger,
	)

	categorySvc := service.NewCategoryService(categoryRepo)

	addressSvc := service.NewAddressService(addressRepo) // ✅ ADDED

	paymentSvc := service.NewPaymentService(
		paymentRepo,
		orderRepo,
		repoLogger,
	)

	// ------------------------------------------------
	// Handlers
	// ------------------------------------------------
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	adminHandler := handler.NewAdminUserHandler(userSvc, orderSvc)
	productHandler := handler.NewProductHandler(productSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	cartHandler := handler.NewCartHandler(cartSvc)
	wishlistHandler := handler.NewWishlistHandler(wishlistSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	addressHandler := handler.NewAddressHandler(addressSvc) // ✅ ADDED
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	// ------------------------------------------------
	// Container
	// ------------------------------------------------
	return &Container{
		UserRepo: userRepo,

		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		AdminHandler:    adminHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
		CartHandler:     cartHandler,
		WishlistHandler: wishlistHandler,
		CategoryHandler: categoryHandler,
		AddressHandler:  addressHandler, // ✅ ADDED
		PaymentHandler:  paymentHandler,

		DBCleanup: sqlDB.Close,
	}, nil
}
