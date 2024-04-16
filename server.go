package main

import (
	serverConfig "dogeplus-backend/config"
	"dogeplus-backend/database"
	"dogeplus-backend/router"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	// load env vars
	config := serverConfig.LoadConfig()

	// define new fiber app and initialize router
	app := router.NewFiberApp()
	router.SetupRoutes(app)

	// Acquire db instance
	dbconn, err := database.GetInstance(config)
	if err != nil {
		log.Fatalf("Failed to establish a DB connection: %v", err)
	}

	// Start server listener loop
	app.Listen(":" + serverConfig.GetEnvWithFallback(config, serverConfig.Port))
}
