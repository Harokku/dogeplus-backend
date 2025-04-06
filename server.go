package main

import (
	"dogeplus-backend/broadcast"
	serverConfig "dogeplus-backend/config"
	"dogeplus-backend/database"
	"dogeplus-backend/router"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// load env vars
	config := serverConfig.LoadConfig()

	// Get db instance
	db, err := database.GetInstance(config)
	if err != nil {
		log.Fatal(err)
	}

	// Init db repos
	repos := database.NewRepositories(db)

	// Init connection manager for realtime broadcast
	connectionManager := broadcast.NewConnectionManager()

	// Init Fiber app
	app := router.NewFiberApp()

	// Enable CORS middleware with default settings
	app.Use(cors.New())

	// Setup routes
	router.SetupRoutes(app, config, repos, connectionManager)

	// Start server listener loop
	app.Listen(":" + serverConfig.GetEnvWithFallback(config, serverConfig.Port))
}
