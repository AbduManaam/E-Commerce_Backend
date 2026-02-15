

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
		logging.Logger.Error("server container build failed", "error", err.Error())
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
	
	
	if cfg.Environment == "development" {
		app.Use(middleware.DebugLogger())
	} else {
		app.Use(middleware.RequestLogger())
	}


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
		HomeHandler: container.HomeHandler,

	}

	routes.SetUpRoutes(app, deps)

	logging.LogInfo(
		"server initialized",
		nil,
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