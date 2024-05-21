package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"time"
)

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:           "DogePlus Backend",
		BodyLimit:         1 * 1024 * 1024,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		EnablePrintRoutes: true,
	})

	// Enable logging
	app.Use(logger.New())

	// Enable CORS
	app.Use(cors.New(cors.ConfigDefault))

	// Enable recover to avoid panics in handler failure
	app.Use(recover.New(recover.ConfigDefault))
	return app
}
