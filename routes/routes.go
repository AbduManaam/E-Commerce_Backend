

package routes

import (
	"backend/config"
	"backend/handler"
	"backend/middleware"
	"backend/repository"

	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type Dependencies struct {
	Logger *slog.Logger
	Cfg    *config.AppConfig

	// Repositories
	UserRepo    repository.UserRepository
	OrderRepo   repository.OrderRepository
	PaymentRepo repository.PaymentRepository

	// Handlers
	AuthHandler      *handler.AuthHandler
	HomeHandler *handler.HomeHandler
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

	authMiddleware := middleware.AuthMiddleware(deps.Cfg, deps.UserRepo)
	adminMiddleware := middleware.AdminMiddleware()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "backend",
		})
	})

	app.Get("/home", deps.HomeHandler.GetHome)

 	auth := app.Group("/auth")
	auth.Post("/signup", deps.AuthHandler.Signup)
	auth.Post("/login", deps.AuthHandler.Login)
	auth.Post("/refresh", deps.AuthHandler.RefreshToken)
	auth.Post("/verify-otp", deps.AuthHandler.VerifyOTP)
	auth.Post("/forgot-password", deps.AuthHandler.ForgotPassword)
	auth.Post("/reset-password", deps.AuthHandler.ResetPasswordWithOTP)
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
	addresses.Get("", deps.AddressHandler.List)
	addresses.Get("/:id", deps.AddressHandler.GetByID)
	addresses.Put("/:id", deps.AddressHandler.Update)
	addresses.Delete("/:id", deps.AddressHandler.Delete)
	addresses.Put("/:id/set-default", deps.AddressHandler.SetDefault)

	// Cart
	cart := api.Group("/cart")
	cart.Post("", deps.CartHandler.Add)
	cart.Get("", deps.CartHandler.Get)
	cart.Put("/item/:itemId", deps.CartHandler.Update)
	cart.Delete("/item/:itemId", deps.CartHandler.Delete)

	// Wishlist
	wishlist := api.Group("/wishlist")
	wishlist.Post("", deps.WishlistHandler.Add)
	wishlist.Get("", deps.WishlistHandler.Get)
	wishlist.Delete("/:product_id", deps.WishlistHandler.Delete)

	// Orders
	orders := api.Group("/orders")
	orders.Post("", deps.OrderHandler.CreateOrder)
	orders.Get("", deps.OrderHandler.GetUserOrders)
	orders.Get("/:id", deps.OrderHandler.GetOrder)
	orders.Put("/:id/cancel", deps.OrderHandler.CancelOrder)

	orderItems := orders.Group("/:order_id/items")

    orderItems.Get("/", deps.OrderHandler.ListOrderItems)
    orderItems.Put("/:item_id/cancel", deps.OrderHandler.CancelOrderItem)


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
   
	// Cloudinary
	admin.Post("/products/:id/image", deps.ProductHandler.UploadImage)
	admin.Delete("/products/images/:id", deps.ProductHandler.DeleteProductImage)
    
	//Refund
	admin.Post("/payments/refund/:order_id", deps.PaymentHandler.RefundPayment)

}
