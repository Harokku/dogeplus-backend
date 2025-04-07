// Package main is the entry point for the DogePlus Backend application.
// It initializes all necessary components including configuration, database,
// repositories, and the web server, then starts the application.
package main

import (
	"dogeplus-backend/broadcast"
	serverConfig "dogeplus-backend/config"
	"dogeplus-backend/database"
	"dogeplus-backend/router"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// main initializes and starts the DogePlus Backend application.
// It sets up all necessary components in the following order:
// 1. Configuration loading
// 2. Database connection
// 3. Repository initialization
// 4. Real-time broadcast manager
// 5. Web server with routes and middleware
// 6. Server startup on the configured port
func main() {
	// Load configuration from environment variables and config files
	config := serverConfig.LoadConfig()

	// Initialize database connection using the loaded configuration
	db, err := database.GetInstance(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create repository instances for database operations
	repos := database.NewRepositories(db)

	// Initialize the connection manager for real-time event broadcasting
	connectionManager := broadcast.NewConnectionManager()

	// Create a new Fiber application instance for HTTP handling
	app := router.NewFiberApp()

	// Enable CORS middleware to allow cross-origin requests
	app.Use(cors.New())

	// Configure all API routes with their respective handlers
	router.SetupRoutes(app, config, repos, connectionManager)

	// Start the HTTP server on the configured port
	app.Listen(":" + serverConfig.GetEnvWithFallback(config, serverConfig.Port))
}
