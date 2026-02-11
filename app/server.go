// package app

// import (
// 	"backend/config"
// 	"backend/middleware"
// 	"backend/routes"
// 	"backend/utils/logging"
// 	"context"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"time"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// )

// type Server struct {
// 	app     *fiber.App
// 	cfg     *config.AppConfig
// 	cleanup func() error
// }

// func NewServer(cfg *config.AppConfig) (*Server, func() error) {
// 	container, err := BuildContainer(cfg)
// 	if err != nil {
// 		logging.LogWarn("server container build failed", nil, err)
// 		os.Exit(1)
// 	}

// 	app := fiber.New(fiber.Config{
// 		AppName:               "Backend API",
// 		CaseSensitive:         true,
// 		StrictRouting:         true,
// 		DisableStartupMessage: true,
// 	})

// 	// ------------------------------------------------
// 	// Middlewares
// 	// ------------------------------------------------
// 	app.Use(middleware.RecoveryMiddleware())
// 	app.Use(cors.New())

// 	app.Use(func(c *fiber.Ctx) error {
// 		start := time.Now()
// 		err := c.Next()
// 		duration := time.Since(start)

// 		logging.LogInfo(
// 			"http request",
// 			c,
// 			"method", c.Method(),
// 			"path", c.Path(),
// 			"status", c.Response().StatusCode(),
// 			"duration_ms", duration.Milliseconds(),
// 		)
// 		return err
// 	})

// 	// ------------------------------------------------
// 	// Health
// 	// ------------------------------------------------
// 	app.Get("/health", func(c *fiber.Ctx) error {
// 		return c.JSON(fiber.Map{"status": "ok"})
// 	})

// 	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
// 		return c.SendStatus(fiber.StatusNoContent)
// 	})

// 	// ------------------------------------------------
// 	// Routes (FIXED)
// 	// ------------------------------------------------
// 	routes.SetUpRoutes(
// 		app,

// 		// handlers
// 		container.AuthHandler,
// 		container.AdminHandler,
// 		container.UserHandler,
// 		container.ProductHandler,
// 		container.OrderHandler,
// 		container.CartHandler,
// 		container.CategoryHandler,
// 		container.WishlistHandler,
// 		container.AddressHandler, // ✅ CORRECT
// 		container.PaymentHandler,

// 		// infra
// 		cfg,
// 		container.UserRepo,
// 	)

// 	logging.LogInfo(
// 		"server initialized",
// 		nil,
// 		"host", cfg.Server.Host,
// 		"port", cfg.Server.Port,
// 	)

// 	return &Server{
// 		app:     app,
// 		cfg:     cfg,
// 		cleanup: container.DBCleanup,
// 	}, container.DBCleanup
// }

// func (s *Server) Run() error {
// 	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

// 	go func() {
// 		logging.LogInfo("server starting", nil, "addr", addr)
// 		if err := s.app.Listen(addr); err != nil {
// 			logging.LogWarn("server failed to start", nil, err, "addr", addr)
// 			os.Exit(1)
// 		}
// 	}()

// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, os.Interrupt)
// 	<-quit

// 	logging.LogInfo("server shutting down...", nil)

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	if err := s.app.ShutdownWithContext(ctx); err != nil {
// 		logging.LogWarn("server shutdown error", nil, err)
// 	}

// 	if err := s.cleanup(); err != nil {
// 		logging.LogWarn("cleanup failed", nil, err)
// 	} else {
// 		logging.LogInfo("cleanup finished successfully", nil)
// 	}

// 	logging.LogInfo("server stopped gracefully", nil)
// 	return nil
// }

package app

import (
	"backend/config"
	"backend/middleware"
	"backend/routes"
	"backend/utils/logging"
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Server struct {
	app     *fiber.App
	cfg     *config.AppConfig
	cleanup func() error
}


func NewServer(cfg *config.AppConfig) (*Server, func() error) {
	// -----------------------------
	// Build the container
	// -----------------------------
	container, err := BuildContainer(cfg)
	if err != nil {
		logging.Logger.Error("server container build failed", "error", err.Error())
		os.Exit(1)
	}

	// -----------------------------
	// Create Fiber app
	// -----------------------------
	app := fiber.New(fiber.Config{
		AppName:               "Backend API",
		CaseSensitive:         true,
		StrictRouting:         true,
		DisableStartupMessage: true,
	})

	// -----------------------------
	// Middlewares
	// -----------------------------
	app.Use(middleware.RecoveryMiddleware())
	app.Use(cors.New())

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		logging.LogInfo(
			"http request",
			c,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
		)
		return err
	})

	// -----------------------------
	// Health & favicon
	// -----------------------------
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	// -----------------------------
	// Routes (Dependency Injection)
	// -----------------------------
	deps := &routes.Dependencies{
		Logger:           logging.Logger,
		Cfg:              cfg,
		UserRepo:         container.UserRepo,
		OrderRepo:        nil, // fill if you need direct access in DI
		PaymentRepo:      nil, // fill if you need direct access in DI
		AuthHandler:      container.AuthHandler,
		AdminUserHandler: container.AdminHandler,
		UserHandler:      container.UserHandler,
		ProductHandler:   container.ProductHandler,
		OrderHandler:     container.OrderHandler,
		CartHandler:      container.CartHandler,
		CategoryHandler:  container.CategoryHandler,
		WishlistHandler:  container.WishlistHandler,
		AddressHandler:   container.AddressHandler,
		PaymentHandler:   container.PaymentHandler,
	}

	routes.SetUpRoutes(app, deps)

	logging.LogInfo(
		"server initialized",
		nil,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	return &Server{
		app:     app,
		cfg:     cfg,
		cleanup: container.DBCleanup,
	}, container.DBCleanup
}


func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	go func() {
		logging.LogInfo("server starting", nil, "addr", addr)
		if err := s.app.Listen(addr); err != nil {
			logging.LogWarn("server failed to start", nil, err, "addr", addr)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logging.LogInfo("server shutting down...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		logging.LogWarn("server shutdown error", nil, err)
	}

	if err := s.cleanup(); err != nil {
		logging.LogWarn("cleanup failed", nil, err)
	} else {
		logging.LogInfo("cleanup finished successfully", nil)
	}

	logging.LogInfo("server stopped gracefully", nil)
	return nil
}
