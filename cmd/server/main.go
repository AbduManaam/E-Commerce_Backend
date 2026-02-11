package main

import (
	"backend/app"
	"backend/config"
	"backend/utils/logging"
	"log"
	"os"
	"time"
)

func main() {

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	logging.Init(env)
	

	cfg, err := config.LoadConfig("app.yaml")
	if err != nil {
		logging.Logger.Error("config loading failed", "error", err.Error())
		os.Exit(1)
	}

	log.Printf("Server local time: %v", time.Now())
log.Printf("Server UTC time:   %v", time.Now().UTC())


	server, cleanup := app.NewServer(cfg)
	defer cleanup()

	if err := server.Run(); err != nil {
		logging.Logger.Error("server error", "error", err.Error())
		os.Exit(1)
	}
}