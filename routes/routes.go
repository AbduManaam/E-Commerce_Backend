
package routes

import (
	"backend/config"
	"backend/handler"
	"backend/middleware"
	"backend/repository"

	"github.com/gofiber/fiber/v2"
)

func SetUpRoutes(
	app *fiber.App,

	// Handlers
	authHandler *handler.AuthHandler,
	adminUserHandler *handler.AdminUserHandler,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	cartHandler *handler.CartHandler,
	categoryHandler *handler.CategoryHandler,
	wishlistHandler *handler.WishlistHandler,
	addressHandler *handler.AddressHandler,
	paymentHandler *handler.PaymentHandler,

	// Infra
	cfg *config.AppConfig,
	userRepo repository.UserRepository,
) {

	authMiddleware := middleware.AuthMiddleware(cfg, userRepo)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "backend",
		})
	})

	auth := app.Group("/auth")
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/verify-otp", authHandler.VerifyOTP)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password-with-otp", authHandler.ResetPasswordWithOTP)
	auth.Post("/resend-verification", authHandler.ResendVerification)

	auth.Post("/change-password", authMiddleware, authHandler.ChangePassword)
	auth.Post("/logout", authMiddleware, authHandler.Logout)

	user := app.Group("/user", authMiddleware)
	user.Get("/profile", userHandler.GetProfile)
	user.Put("/profile", userHandler.UpdateProfile)

	products := app.Group("/products")
	products.Get("/", productHandler.ListProducts)
	products.Get("/filter", productHandler.ListFiltered)
	products.Get("/:id", productHandler.GetProduct)

	categories := app.Group("/categories")
	categories.Get("/", categoryHandler.List)


	api := app.Group("/api", authMiddleware)

	addresses := api.Group("/addresses")
	addresses.Post("", addressHandler.Create)
	addresses.Get("/", addressHandler.List)
	addresses.Get("/:id", addressHandler.GetByID)
	addresses.Put("/:id", addressHandler.Update)
	addresses.Delete("/:id", addressHandler.Delete)
	addresses.Put("/:id/set-default", addressHandler.SetDefault)

	cart := api.Group("/cart")
	cart.Post("/", cartHandler.Add)
	cart.Get("/", cartHandler.Get)
	cart.Put("/item/:itemId", cartHandler.Update)
	cart.Delete("/item/:itemId", cartHandler.Delete)

	wishlist := api.Group("/wishlist")
	wishlist.Post("/", wishlistHandler.Add)
	wishlist.Get("/", wishlistHandler.Get)
	wishlist.Delete("/:product_id", wishlistHandler.Delete)

	orders := api.Group("/orders")
	orders.Post("/", orderHandler.CreateOrder)
	orders.Get("/", orderHandler.GetUserOrders)
	orders.Get("/:id", orderHandler.GetOrder)
	orders.Put("/:id/cancel", orderHandler.CancelOrder)

	payments := api.Group("/payments")
	payments.Post("/intent", paymentHandler.CreatePaymentIntent)

	
	admin := app.Group(
		"/admin",
		authMiddleware,
		middleware.AdminMiddleware(),
	)

	admin.Get("/users", adminUserHandler.ListUsers)
	admin.Put("/users/:id", adminUserHandler.UpdateUser)
	admin.Put("/users/:id/block", adminUserHandler.BlockUser)
	admin.Get("/users/:user_id/orders", adminUserHandler.GetUserOrders)

	admin.Get("/orders", orderHandler.ListAllOrders)
	admin.Put("/orders/:id/status", orderHandler.UpdateOrderStatus)
	admin.Put("/orders/:id", orderHandler.AdminUpdateOrder)

	// Admin products
	admin.Post("/products", productHandler.CreateProduct)
	admin.Put("/products/:id", productHandler.UpdateProduct)
	admin.Delete("/products/:id", productHandler.DeleteProduct)

	// Admin categories
	admin.Post("/categories", categoryHandler.Create)
}
