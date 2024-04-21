package router

import (
	"dogeplus-backend/database"
	"dogeplus-backend/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, repos *database.Repositories) {
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
	v1.Post("/active-events", handlers.CreateNewEvent(repos))
	v1.Get("/active-events", handlers.GetSingleEvent(repos))
	v1.Put("/active-events", handlers.UpdateEventTask(repos))
}
