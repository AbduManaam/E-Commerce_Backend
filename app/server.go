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
)

type Server struct {
	app     *fiber.App
	cfg     *config.AppConfig
	cleanup func() error
}

func NewServer(cfg *config.AppConfig) (*Server, func() error) {

	// Build the container
	container, err := BuildContainer(cfg)
	if err != nil {
		logging.LogError("server container build failed", "error", err.Error())
		os.Exit(1)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Backend API",
		CaseSensitive:         true,
		StrictRouting:         true,
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
	})

	app.Use(middleware.CORSMiddleware())

	app.Use(middleware.RecoveryMiddleware())

	// Single consolidated request logger for all environments.
	// Dev vs prod differentiation happens at the logger output level (text vs JSON).
	app.Use(middleware.RequestLogger())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Get("/test-cors", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "CORS is working!",
			"origin":  c.Get("Origin"),
			"time":    time.Now().UTC(),
		})
	})

	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Routes (Dependency Injection)
	deps := &routes.Dependencies{
		Logger:           logging.Logger,
		Cfg:              cfg,
		UserRepo:         container.UserRepo,
		OrderRepo:        nil,
		PaymentRepo:      nil,
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
		HomeHandler:      container.HomeHandler,
	}

	routes.SetUpRoutes(app, deps)

	logging.LogInfo("server initialized",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"environment", cfg.Environment,
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
		logging.LogInfo("server starting", "addr", addr)
		if err := s.app.Listen(addr); err != nil {
			logging.LogError("server failed to start", "error", err.Error(), "addr", addr)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logging.LogInfo("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		logging.LogError("server shutdown error", "error", err.Error())
	}

	if err := s.cleanup(); err != nil {
		logging.LogError("cleanup failed", "error", err.Error())
	} else {
		logging.LogInfo("cleanup finished successfully")
	}

	logging.LogInfo("server stopped gracefully")
	return nil
}
