package router

import (
	"dogeplus-backend/broadcast"
	"dogeplus-backend/config"
	"dogeplus-backend/database"
	"dogeplus-backend/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up the routes for the application.
// It takes in the following parameters:
// - app: a pointer to a fiber.App instance representing the application
// - config: a config.Config struct with env variables
// - repos: a pointer to a database.Repositories instance representing the collection of repositories
// - cm: a pointer to a broadcast.ConnectionManager instance representing the connection manager
func SetupRoutes(app *fiber.App, config config.Config, repos *database.Repositories, cm *broadcast.ConnectionManager) {
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
	tasks.Get("/", handlers.GetTasks(config, repos))
	tasks.Post("/", handlers.GetTasksForEscalation(repos))
	tasks.Post("/upload/main", handlers.UploadMainTasksFile(repos))
	tasks.Post("/upload/local", handlers.UploadLocalTasksFile(config))

	// ActiveEvents routes
	activeEvents := v1.Group("/active-events")
	activeEvents.Post("/", handlers.CreateNewEvent(repos, config))
	activeEvents.Post("/overview", handlers.PostNewOverview(repos, cm))
	activeEvents.Put("/", handlers.UpdateEventTask(repos, cm))
	activeEvents.Get("/:central_id", handlers.GetSingleEvent(repos))
	activeEvents.Get("/:central_id/:event_nr", handlers.GetSpecificEvent(repos))
	//activeEvents.Get("/aggregated_status", )

	// Event aggregation routes
	completionAggregation := v1.Group("/completion_aggregation")
	completionAggregation.Get("/", handlers.GetAllTaskCompletionInfo(cm))
	completionAggregation.Get("/:event_number", handlers.GetTaskCompletionInfoForKey(cm))

	// Event Escalation routes
	aggregationEscalation := v1.Group("/escalation_aggregation")
	aggregationEscalation.Get("/", handlers.GetAllEscalationLevels)
	aggregationEscalation.Get("/details", handlers.GetAllEscalationDetails(repos))
	aggregationEscalation.Get("/details/:central_id", handlers.GetEscalationDetailsByCentralId(repos))
	aggregationEscalation.Get("/details/:central_id/:event_number", handlers.GetEscalationDetailsByCentralIdAndEventNumber(repos))
	aggregationEscalation.Post("/escalate", handlers.PostEscalate(repos, config, cm))
	aggregationEscalation.Post("/deescalate", handlers.PostDeEscalate(repos, config, cm))

	// Escalation Levels Definitions
	escalationLevels := v1.Group("/escalation_levels")
	escalationLevels.Get("/", handlers.GetAllEscalationLevelsDefinitions(repos))

	// Ws Routes
	websocket := v1.Group("/ws")
	websocket.Get("/", handlers.WsUpgrader(cm), handlers.WsHandler(cm))
}
