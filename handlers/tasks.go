package handlers

import (
	"dogeplus-backend/database"
	"github.com/gofiber/fiber/v2"
)

type EscalationRequest struct {
	Category   string `json:"category"`
	StartLevel string `json:"start_level"`
	EndLevel   string `json:"end_level"`
}

// GetTasks returns a handler function that retrieves distinct categories from the database
// and sends them as a response in JSON format.
// The handler function takes a *fiber.Ctx as input and returns an error.
// It calls the GetCategories method of the TaskRepository struct in the *database.Repositories instance
// to retrieve the distinct categories from the "tasks" table in the database.
// If an error occurs during the retrieval, an error response is sent with the status code 500
// and a JSON object containing the error details.
// Otherwise, a success response is sent with the status code 200
// and a JSON object with the retrieved categories, the length of the categories list,
// and a result message.
func GetTasks(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		// Get distinct categories slice from db
		categoriesList, err := repos.Tasks.GetCategories()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to get categories",
				"detail": err.Error(),
			})
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"result": "Retrieved categories",
			"length": len(categoriesList),
			"data":   categoriesList,
		})
	}
}

// GetTasksForEscalation handles the retrieval of tasks for escalation based on category and escalation levels.
func GetTasksForEscalation(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {

		// Create a new instance of EscalationRequest
		var escalationRequest EscalationRequest

		// Parse the JSON body into the escalationRequest instance
		if err := ctx.BodyParser(&escalationRequest); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Failed to parse request body",
				"detail": err.Error(),
			})
		}

		// Get missing tasks fro db
		categoriesList, err := repos.Tasks.GetGyCategoryAndEscalationLevel(escalationRequest.Category, escalationRequest.StartLevel, escalationRequest.EndLevel)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to get tasks",
				"detail": err.Error(),
			})
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"result": "Retrieved tasks",
			"length": len(categoriesList),
			"data":   categoriesList,
		})
	}
}
