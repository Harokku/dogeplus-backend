package main

import "dogeplus-backend/router"

func main() {
	// load env vars
	config := LoadConfig()

	// define new fiber app and initialize router
	app := router.NewFiberApp()
	router.SetupRoutes(app)

	// Start server listener loop
	app.Listen(":" + GetEnvWithFallback(config, Port))
}
