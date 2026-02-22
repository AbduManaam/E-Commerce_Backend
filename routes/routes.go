

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
	userOnlyMiddleware := middleware.UserOnlyMiddleware()

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

	// Addresses (mutations require user role; admin read-only)
	addresses := api.Group("/addresses")
	addresses.Post("", userOnlyMiddleware, deps.AddressHandler.Create)
	addresses.Get("", deps.AddressHandler.List)
	addresses.Get("/:id", deps.AddressHandler.GetByID)
	addresses.Put("/:id", userOnlyMiddleware, deps.AddressHandler.Update)
	addresses.Delete("/:id", userOnlyMiddleware, deps.AddressHandler.Delete)
	addresses.Put("/:id/set-default", userOnlyMiddleware, deps.AddressHandler.SetDefault)

	// Cart (mutations require user role; admin read-only)
	cart := api.Group("/cart")
	cart.Post("", userOnlyMiddleware, deps.CartHandler.Add)
	cart.Get("", deps.CartHandler.Get)
	cart.Put("/item/:itemId", userOnlyMiddleware, deps.CartHandler.Update)
	cart.Delete("/item/:itemId", userOnlyMiddleware, deps.CartHandler.Delete)

	// Wishlist (mutations require user role; admin read-only)
	wishlist := api.Group("/wishlist")
	wishlist.Post("", userOnlyMiddleware, deps.WishlistHandler.Add)
	wishlist.Get("", deps.WishlistHandler.Get)
	wishlist.Delete("/:product_id", userOnlyMiddleware, deps.WishlistHandler.Delete)

	// Orders (mutations require user role; admin read-only)
	orders := api.Group("/orders")
	orders.Post("", userOnlyMiddleware, deps.OrderHandler.CreateOrder)
	orders.Get("", deps.OrderHandler.GetUserOrders)
	orders.Get("/:id", deps.OrderHandler.GetOrder)
	orders.Put("/:id/cancel", userOnlyMiddleware, deps.OrderHandler.CancelOrder)

	orderItems := orders.Group("/:order_id/items")
	orderItems.Get("/", deps.OrderHandler.ListOrderItems)
	orderItems.Put("/:item_id/cancel", userOnlyMiddleware, deps.OrderHandler.CancelOrderItem)


	// Payments (mutations require user role; admin read-only)
	payments := api.Group("/payments")
	payments.Post("/intent", userOnlyMiddleware, deps.PaymentHandler.CreatePaymentIntent)
	payments.Post("/confirm", userOnlyMiddleware, deps.PaymentHandler.ConfirmPayment)

	// ----------------- Admin -----------------
	admin := app.Group("/admin", authMiddleware, adminMiddleware)

	// Users
	admin.Get("/users", deps.AdminUserHandler.ListUsers)
	admin.Put("/users/:id", deps.AdminUserHandler.UpdateUser)
	admin.Put("/users/:id/block", deps.AdminUserHandler.BlockUser)
	admin.Get("/users/:user_id/orders", deps.AdminUserHandler.GetUserOrders)
	admin.Put("/users/:id/unblock", deps.AdminUserHandler.UnblockUser)

	// Orders
	admin.Get("/orders", deps.OrderHandler.ListAllOrders)
	admin.Put("/orders/:id/status", deps.OrderHandler.UpdateOrderStatus)
	admin.Put("/orders/:id", deps.OrderHandler.AdminUpdateOrder)
	admin.Get("/orders/:id", deps.OrderHandler.AdminGetOrder)

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
