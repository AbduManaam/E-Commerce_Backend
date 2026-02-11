// package routes

// import (
// 	"backend/config"
// 	"backend/handler"
// 	"backend/middleware"
// 	"backend/repository"
// 	"backend/service"
// 	"backend/utils/logging"

// 	"github.com/gofiber/fiber/v2"
// )

// func SetUpRoutes(
// 	app *fiber.App,

// 	// Handlers
// 	authHandler *handler.AuthHandler,
// 	adminUserHandler *handler.AdminUserHandler,
// 	userHandler *handler.UserHandler,
// 	productHandler *handler.ProductHandler,
// 	orderHandler *handler.OrderHandler,
// 	cartHandler *handler.CartHandler,
// 	categoryHandler *handler.CategoryHandler,
// 	wishlistHandler *handler.WishlistHandler,
// 	addressHandler *handler.AddressHandler,
// 	paymentHandler *handler.PaymentHandler,

// 	// Infra
// 	cfg *config.AppConfig,
// 	userRepo repository.UserRepository,
// ) {

// 	authMiddleware := middleware.AuthMiddleware(cfg, userRepo)

// 	app.Get("/health", func(c *fiber.Ctx) error {
// 		return c.JSON(fiber.Map{
// 			"status":  "ok",
// 			"service": "backend",
// 		})
// 	})

// 	auth := app.Group("/auth")
// 	auth.Post("/signup", authHandler.Signup)
// 	auth.Post("/login", authHandler.Login)
// 	auth.Post("/refresh", authHandler.RefreshToken)
// 	auth.Post("/verify-otp", authHandler.VerifyOTP)
// 	auth.Post("/forgot-password", authHandler.ForgotPassword)
// 	auth.Post("/reset-password-with-otp", authHandler.ResetPasswordWithOTP)
// 	auth.Post("/resend-verification", authHandler.ResendVerification)

// 	auth.Post("/change-password", authMiddleware, authHandler.ChangePassword)
// 	auth.Post("/logout", authMiddleware, authHandler.Logout)

// 	user := app.Group("/user", authMiddleware)
// 	user.Get("/profile", userHandler.GetProfile)
// 	user.Put("/profile", userHandler.UpdateProfile)

// 	products := app.Group("/products")
// 	products.Get("/", productHandler.ListProducts)
// 	products.Get("/filter", productHandler.ListFiltered)
// 	products.Get("/:id", productHandler.GetProduct)

// 	categories := app.Group("/categories")
// 	categories.Get("/", categoryHandler.List)


// 	api := app.Group("/api", authMiddleware)

// 	addresses := api.Group("/addresses")
// 	addresses.Post("", addressHandler.Create)
// 	addresses.Get("/", addressHandler.List)
// 	addresses.Get("/:id", addressHandler.GetByID)
// 	addresses.Put("/:id", addressHandler.Update)
// 	addresses.Delete("/:id", addressHandler.Delete)
// 	addresses.Put("/:id/set-default", addressHandler.SetDefault)

// 	cart := api.Group("/cart")
// 	cart.Post("/", cartHandler.Add)
// 	cart.Get("/", cartHandler.Get)
// 	cart.Put("/item/:itemId", cartHandler.Update)
// 	cart.Delete("/item/:itemId", cartHandler.Delete)

// 	wishlist := api.Group("/wishlist")
// 	wishlist.Post("/", wishlistHandler.Add)
// 	wishlist.Get("/", wishlistHandler.Get)
// 	wishlist.Delete("/:product_id", wishlistHandler.Delete)

// 	orders := api.Group("/orders")
// 	orders.Post("/", orderHandler.CreateOrder)
// 	orders.Get("/", orderHandler.GetUserOrders)
// 	orders.Get("/:id", orderHandler.GetOrder)
// 	orders.Put("/:id/cancel", orderHandler.CancelOrder)

	
// 	paymentService := service.NewPaymentService(
//     paymentRepo,
//     orderRepo,
//     logging.Logger, // your logging package instance
// )
	
// 	paymentHandler = handler.NewPaymentHandler(paymentService)
// 	payments := api.Group("/payments")

// 	// Create handler instance

// 	// Payment endpoints
// 	payments.Post("/intent", paymentHandler.CreatePaymentIntent)
// 	payments.Post("/confirm", paymentHandler.ConfirmPayment)


	
// 	admin := app.Group(
// 		"/admin",
// 		authMiddleware,
// 		middleware.AdminMiddleware(),
// 	)

// 	admin.Get("/users", adminUserHandler.ListUsers)
// 	admin.Put("/users/:id", adminUserHandler.UpdateUser)
// 	admin.Put("/users/:id/block", adminUserHandler.BlockUser)
// 	admin.Get("/users/:user_id/orders", adminUserHandler.GetUserOrders)

// 	admin.Get("/orders", orderHandler.ListAllOrders)
// 	admin.Put("/orders/:id/status", orderHandler.UpdateOrderStatus)
// 	admin.Put("/orders/:id", orderHandler.AdminUpdateOrder)

// 	// Admin products
// 	admin.Post("/products", productHandler.CreateProduct)
// 	admin.Put("/products/:id", productHandler.UpdateProduct)
// 	admin.Delete("/products/:id", productHandler.DeleteProduct)

// 	// Admin categories
// 	admin.Post("/categories", categoryHandler.Create)
// }


package routes

import (
	"backend/config"
	"backend/handler"
	"backend/middleware"
	"backend/repository"

	"github.com/gofiber/fiber/v2"
	"log/slog"
)

// Dependencies aggregates all services, handlers, and repositories for injection.
type Dependencies struct {
	Logger *slog.Logger
	Cfg    *config.AppConfig

	// Repositories
	UserRepo    repository.UserRepository
	OrderRepo   repository.OrderRepository
	PaymentRepo repository.PaymentRepository

	// Handlers
	AuthHandler      *handler.AuthHandler
	AdminUserHandler *handler.AdminUserHandler
	UserHandler      *handler.UserHandler
	ProductHandler   *handler.ProductHandler
	OrderHandler     *handler.OrderHandler
	CartHandler      *handler.CartHandler
	CategoryHandler  *handler.CategoryHandler
	WishlistHandler  *handler.WishlistHandler
	AddressHandler   *handler.AddressHandler
	PaymentHandler   *handler.PaymentHandler
}

func SetUpRoutes(app *fiber.App, deps *Dependencies) {
	// ----------------- Middleware -----------------
	authMiddleware := middleware.AuthMiddleware(deps.Cfg, deps.UserRepo)
	adminMiddleware := middleware.AdminMiddleware()

	// ----------------- Health Check -----------------
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "backend",
		})
	})

	// ----------------- Auth -----------------
	auth := app.Group("/auth")
	auth.Post("/signup", deps.AuthHandler.Signup)
	auth.Post("/login", deps.AuthHandler.Login)
	auth.Post("/refresh", deps.AuthHandler.RefreshToken)
	auth.Post("/verify-otp", deps.AuthHandler.VerifyOTP)
	auth.Post("/forgot-password", deps.AuthHandler.ForgotPassword)
	auth.Post("/reset-password-with-otp", deps.AuthHandler.ResetPasswordWithOTP)
	auth.Post("/resend-verification", deps.AuthHandler.ResendVerification)
	auth.Post("/change-password", authMiddleware, deps.AuthHandler.ChangePassword)
	auth.Post("/logout", authMiddleware, deps.AuthHandler.Logout)

	// ----------------- User -----------------
	user := app.Group("/user", authMiddleware)
	user.Get("/profile", deps.UserHandler.GetProfile)
	user.Put("/profile", deps.UserHandler.UpdateProfile)

	// ----------------- Products & Categories -----------------
	products := app.Group("/products")
	products.Get("/", deps.ProductHandler.ListProducts)
	products.Get("/filter", deps.ProductHandler.ListFiltered)
	products.Get("/:id", deps.ProductHandler.GetProduct)

	categories := app.Group("/categories")
	categories.Get("/", deps.CategoryHandler.List)

	// ----------------- API Routes (Authenticated) -----------------
	api := app.Group("/api", authMiddleware)

	// Addresses
	addresses := api.Group("/addresses")
	addresses.Post("", deps.AddressHandler.Create)
	addresses.Get("/", deps.AddressHandler.List)
	addresses.Get("/:id", deps.AddressHandler.GetByID)
	addresses.Put("/:id", deps.AddressHandler.Update)
	addresses.Delete("/:id", deps.AddressHandler.Delete)
	addresses.Put("/:id/set-default", deps.AddressHandler.SetDefault)

	// Cart
	cart := api.Group("/cart")
	cart.Post("/", deps.CartHandler.Add)
	cart.Get("/", deps.CartHandler.Get)
	cart.Put("/item/:itemId", deps.CartHandler.Update)
	cart.Delete("/item/:itemId", deps.CartHandler.Delete)

	// Wishlist
	wishlist := api.Group("/wishlist")
	wishlist.Post("/", deps.WishlistHandler.Add)
	wishlist.Get("/", deps.WishlistHandler.Get)
	wishlist.Delete("/:product_id", deps.WishlistHandler.Delete)

	// Orders
	orders := api.Group("/orders")
	orders.Post("/", deps.OrderHandler.CreateOrder)
	orders.Get("/", deps.OrderHandler.GetUserOrders)
	orders.Get("/:id", deps.OrderHandler.GetOrder)
	orders.Put("/:id/cancel", deps.OrderHandler.CancelOrder)

	// Payments
	payments := api.Group("/payments")
	payments.Post("/intent", deps.PaymentHandler.CreatePaymentIntent)
	payments.Post("/confirm", deps.PaymentHandler.ConfirmPayment)

	// ----------------- Admin -----------------
	admin := app.Group("/admin", authMiddleware, adminMiddleware)

	// Users
	admin.Get("/users", deps.AdminUserHandler.ListUsers)
	admin.Put("/users/:id", deps.AdminUserHandler.UpdateUser)
	admin.Put("/users/:id/block", deps.AdminUserHandler.BlockUser)
	admin.Get("/users/:user_id/orders", deps.AdminUserHandler.GetUserOrders)

	// Orders
	admin.Get("/orders", deps.OrderHandler.ListAllOrders)
	admin.Put("/orders/:id/status", deps.OrderHandler.UpdateOrderStatus)
	admin.Put("/orders/:id", deps.OrderHandler.AdminUpdateOrder)

	// Products
	admin.Post("/products", deps.ProductHandler.CreateProduct)
	admin.Put("/products/:id", deps.ProductHandler.UpdateProduct)
	admin.Delete("/products/:id", deps.ProductHandler.DeleteProduct)

	// Categories
	admin.Post("/categories", deps.CategoryHandler.Create)
}
