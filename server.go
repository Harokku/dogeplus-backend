package main

import "github.com/gofiber/fiber/v3"

func main() {
	// load env vars
	config := LoadConfig()

	var app *fiber.App

	// Start server listener loop
	err := app.Listen(GetEnvWithFallback(config, Port))
	if err != nil {
		return
	}
}
