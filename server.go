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

	// Get db instance
	db, err := database.GetInstance(config)
	if err != nil {
		log.Fatal(err)
	}

	// Init db repos
	repos := database.NewRepositories(db)

	// Init Fiber app
	app := router.NewFiberApp()

	// Setup routes
	router.SetupRoutes(app, repos)

	// Start server listener loop
	app.Listen(":" + serverConfig.GetEnvWithFallback(config, serverConfig.Port))
}
