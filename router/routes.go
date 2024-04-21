package router

import (
	"dogeplus-backend/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/", handlers.HomeHandler)

	// api group
	api := app.Group("/api")

	// V1 api group
	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("Version", "v1")
		return c.Next()
	})
	v1.Get("/", handlers.VersionLandingHandler("version 1"))

	//TODO: Tasks routes

	//TODO: ActiveEvents routes
}
