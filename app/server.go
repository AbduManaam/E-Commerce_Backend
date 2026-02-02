package app

import (
	"backend/config"
	"backend/middleware"
	"backend/routes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	app     *fiber.App
	cfg     *config.AppConfig
	cleanup func() error
}

func NewServer(cfg *config.AppConfig) (*Server, func()error) {
	container, err := BuildContainer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		AppName:               "Backend API",
		CaseSensitive:         true,
		StrictRouting:         true,
		DisableStartupMessage: true,
	})

	// middlewares
	app.Use(middleware.RecoveryMiddleware())
	app.Use(cors.New())
	app.Use(logger.New())

	// health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
    return c.SendStatus(fiber.StatusNoContent)
})

	// routes
	routes.SetUpRoutes(
		app,
		container.AuthHandler,
		container.AdminHandler,
		container.UserHandler,
		container.ProductHandler,
		container.OrderHandler,
		container.CartHandler,
		container.WishlistHandler,
		cfg,
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
		log.Printf("server running at %s", addr)
		if err := s.app.Listen(addr); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return err
	}

	return s.cleanup()
}











