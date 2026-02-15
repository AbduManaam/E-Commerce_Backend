package app

import (
	"backend/config"
	"backend/handler"
	"backend/repository"
	"backend/service"
	"backend/utils/databases"
	"backend/utils/email"
	"backend/utils/logging"
	cloudinaryutil "backend/utils/utils/cloudinary"
	"log"
	"os"
)

type Container struct {
	// Repositories
	UserRepo repository.UserRepository

	// Handlers
	AuthHandler     *handler.AuthHandler
	HomeHandler     *handler.HomeHandler
	UserHandler     *handler.UserHandler
	AdminHandler    *handler.AdminUserHandler
	ProductHandler  *handler.ProductHandler
	OrderHandler    *handler.OrderHandler
	CartHandler     *handler.CartHandler
	WishlistHandler *handler.WishlistHandler
	CategoryHandler *handler.CategoryHandler
	AddressHandler  *handler.AddressHandler
	PaymentHandler  *handler.PaymentHandler

	DBCleanup func() error
}

func BuildContainer(cfg *config.AppConfig) (*Container, error) {

	// ---------------- LOGGER ----------------
	logging.Init(cfg.Environment)
	logger := logging.Logger

	// ---------------- DATABASE ----------------
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

	// ---------------- REPOSITORIES ----------------
	authRepo := repository.NewAuthRepository(db)
	userRepo := repository.NewUserRepository(db, logger)
	productRepo := repository.NewProductRepository(db, logger)
	orderRepo := repository.NewOrderRepository(db, logger)
	wishlistRepo := repository.NewWishlistRepository(db, logger)
	categoryRepo := repository.NewCategoryRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	heroRepo := repository.NewHeroRepo()
	featureRepo := repository.NewFeatureRepo()
	reviewRepo := repository.NewReviewRepo()

	var cartRepo repository.CartRepositoryInterface =
		repository.NewCartRepository(db, logger)

	var productReader repository.ProductReader = productRepo
	var productWriter repository.ProductWriter = productRepo

	// ---------------- EMAIL ----------------
	email.Init(cfg.SMTP)
	emailSvc := email.NewSMTPService()

	// ---------------- CLOUDINARY ----------------
	cloudinaryClient, err := cloudinaryutil.New(
		cfg.Cloudinary.CloudName,
		cfg.Cloudinary.APIKey,
		cfg.Cloudinary.APISecret,
	)
	if err != nil {
		log.Fatal(err)
	}

	// ---------------- SERVICES ----------------
	homeSvc := service.NewHomeService(
		heroRepo,
		productRepo,
		featureRepo,
		reviewRepo,
	)

	authSvc := service.NewAuthService(
		userRepo,
		authRepo,
		&cfg.JWT,
		emailSvc,
	)

	userSvc := service.NewUserService(
		userRepo,
		authRepo,
		logger,
	)

	productSvc := service.NewProductService(
		db,
		productRepo,
		cloudinaryClient,
		logger,
	)

	orderSvc := service.NewOrderService(
		orderRepo,
		productReader,
		productWriter,
		cartRepo,
		addressRepo,
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
	addressSvc := service.NewAddressService(addressRepo)

	paymentSvc := service.NewPaymentService(
		paymentRepo,
		orderRepo,
		repoLogger,
	)

	// ---------------- HANDLERS ----------------
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	adminHandler := handler.NewAdminUserHandler(userSvc, orderSvc)
	productHandler := handler.NewProductHandler(productSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	cartHandler := handler.NewCartHandler(cartSvc)
	wishlistHandler := handler.NewWishlistHandler(wishlistSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	addressHandler := handler.NewAddressHandler(addressSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)
	homeHandler := handler.NewHomeHandler(homeSvc)

	// ---------------- CONTAINER ----------------
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
		AddressHandler:  addressHandler,
		PaymentHandler:  paymentHandler,
		HomeHandler:     homeHandler,

		DBCleanup: sqlDB.Close,
	}, nil
}
