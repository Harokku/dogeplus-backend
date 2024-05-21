package router

import (
	"dogeplus-backend/broadcast"
	"dogeplus-backend/database"
	"dogeplus-backend/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up the routes for the application.
// It takes in the following parameters:
// - app: a pointer to a fiber.App instance representing the application
// - repos: a pointer to a database.Repositories instance representing the collection of repositories
// - cm: a pointer to a broadcast.ConnectionManager instance representing the connection manager
func SetupRoutes(app *fiber.App, repos *database.Repositories, cm *broadcast.ConnectionManager) {
	app.Get("/", handlers.HomeHandler)

	// api group
	api := app.Group("/api")

	// V1 api group
	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("Version", "v1")
		return c.Next()
	})
	v1.Get("/", handlers.VersionLandingHandler("version 1"))

	// Tasks routes
	tasks := v1.Group("/tasks")
	tasks.Get("/", handlers.GetTasks(repos))

	// ActiveEvents routes
	activeEvents := v1.Group("/active-events")
	activeEvents.Post("/", handlers.CreateNewEvent(repos))
	activeEvents.Put("/", handlers.UpdateEventTask(repos))
	activeEvents.Get("/:central_id", handlers.GetSingleEvent(repos))
	activeEvents.Get("/:central_id/:event_nr", handlers.GetSpecificEvent(repos))

	// Ws Routes
	websocket := v1.Group("/ws")
	websocket.Get("/", handlers.WsUpgrader(cm), handlers.WsHandler(cm))
}
