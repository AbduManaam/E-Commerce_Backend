package routes

import (
	"backend/config"
	"backend/handler"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetUpRoutes(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	adminUserHandler *handler.AdminUserHandler,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	cartHandler *handler.CartHandler,
	categoryHandler *handler.CategoryHandler,
	wishlistHandler *handler.WishlistHandler,
	cfg *config.AppConfig,
) {
	// Public routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "backend"})
	})

	// Auth routes
	auth := app.Group("/auth")
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/verify-otp", authHandler.VerifyOTP)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password-with-otp", authHandler.ResetPasswordWithOTP)
	auth.Post("/change-password", middleware.AuthMiddleware(cfg), authHandler.ChangePassword)
	auth.Post("/logout", middleware.AuthMiddleware(cfg), authHandler.Logout)

	// Protected user routes
	user := app.Group("/user", middleware.AuthMiddleware(cfg))
	user.Get("/profile", userHandler.GetProfile)
	user.Put("/profile", userHandler.UpdateProfile)

	// Product routes
	products := app.Group("/products")
	products.Get("/", productHandler.ListProducts)
	products.Get("/:id", productHandler.GetProduct)
	products.Get("/filter", productHandler.ListFiltered)


	// Admin routes
	admin := app.Group("/admin", middleware.AuthMiddleware(cfg), middleware.AdminMiddleware())
	admin.Put("/users/:id", adminUserHandler.UpdateUser)
	admin.Put("/users/:id/block", adminUserHandler.BlockUser)
	admin.Post("/products", productHandler.CreateProduct)
	admin.Put("/products/:id", productHandler.UpdateProduct)
	admin.Delete("/products/:id", productHandler.DeleteProduct)
	admin.Post("/categories",categoryHandler.Create )

    categories:= app.Group("/categories")
	categories.Get("/",categoryHandler.List)


	// Order routes
	orders := app.Group("/orders", middleware.AuthMiddleware(cfg))
	orders.Post("/", orderHandler.CreateOrder)
	orders.Get("/", orderHandler.GetUserOrders)
	orders.Get("/:id", orderHandler.GetOrder)
	orders.Put("/:id/cancel", orderHandler.CancelOrder)
	admin.Get("/orders", orderHandler.ListAllOrders)
	admin.Put("/orders/:id/status", orderHandler.UpdateOrderStatus)
	admin.Put("/order/:id", orderHandler.AdminUpdateOrder)

	api := app.Group("/api")
	authMiddleware := middleware.AuthMiddleware(cfg)

	// CART ROUTES 
	cart := api.Group("/cart", authMiddleware)
	cart.Post("/", cartHandler.Add)   
	cart.Get("/", cartHandler.Get)
	cart.Put("/item/:itemId", cartHandler.Update)
	cart.Delete("/item/:itemId", cartHandler.Delete)

	// WISHLIST ROUTES 
	wishlist := api.Group("/wishlist", authMiddleware)
	wishlist.Post("/", wishlistHandler.Add)             
	wishlist.Get("/", wishlistHandler.Get)
	wishlist.Delete("/:product_id", wishlistHandler.Delete)
}
