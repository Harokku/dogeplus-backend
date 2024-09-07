package handlers

import (
	"dogeplus-backend/database"
	"github.com/gofiber/fiber/v2"
)

type taskCompletionInfo struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// -------------------------
// Task completion info functions
// -------------------------

//region TaskCompletionInfo

// GetAllTaskCompletionInfo returns a JSON representation of the task completion information for all tasks.
func GetAllTaskCompletionInfo(c *fiber.Ctx) error {
	taskMap := database.GetTaskCompletionMapInstance(nil)

	allTasks := make(map[int]taskCompletionInfo)
	for key, value := range taskMap.Data {
		allTasks[key] = taskCompletionInfo{
			Completed: value.Completed,
			Total:     value.Total,
		}
	}

	return c.JSON(allTasks)
}

// GetTaskCompletionInfoForKey extracts the event number from the request URL parameters and retrieves task completion information.
//
// If the event number is invalid, the function returns a 400 Bad Request response.
// If the event is not found, the function returns a 404 Not Found response.
// If the event is found, the function returns the task completion information as JSON.
func GetTaskCompletionInfoForKey(c *fiber.Ctx) error {
	key, err := c.ParamsInt("event_number")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task key",
		})
	}

	taskMap := database.GetTaskCompletionMapInstance(nil)

	if value, ok := taskMap.Data[key]; ok {
		taskInfo := taskCompletionInfo{
			Completed: value.Completed,
			Total:     value.Total,
		}
		return c.JSON(taskInfo)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "Event not found",
	})
}

//endregion

// -------------------------
// Escalation levels functions
// -------------------------

//region EscalationLevels

// EscalateRequest Request payload structure
type EscalateRequest struct {
	EventNumber int            `json:"eventNumber"`
	NewLevel    database.Level `json:"newLevel"`
}

// GetAllEscalationLevels handles the HTTP request to retrieve all escalation levels.
// It fetches the escalation levels from the database and returns them as a JSON response.
func GetAllEscalationLevels(c *fiber.Ctx) error {
	escalationLevels := database.GetEscalationLevelsInstance(nil)
	levelData := escalationLevels.GetLevels()
	return c.JSON(levelData)
}

// PostEscalate handles the escalation of an event to a new level based on the request payload.
// It parses the request body into an EscalateRequest, retrieves the escalation levels instance,
// and calls the Escalate method to update the event's level if the new level is valid.
// Returns a JSON response indicating success or error based on the outcome.
func PostEscalate(c *fiber.Ctx) error {
	// Parse request body
	var request EscalateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request",
		})
	}

	// Get escalation map instance
	escalationLevels := database.GetEscalationLevelsInstance(nil)

	// Call the Escalate method
	err := escalationLevels.Escalate(request.EventNumber, request.NewLevel)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Event level escalated successfully",
	})
}

// PostDeEscalate handles de-escalation of an event level based on the provided request data and updates the escalation map.
func PostDeEscalate(c *fiber.Ctx) error {
	// Parse request body
	var request EscalateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request",
		})
	}

	// Get escalation map instance
	escalationLevels := database.GetEscalationLevelsInstance(nil)

	// Call the Escalate method
	err := escalationLevels.Deescalate(request.EventNumber, request.NewLevel)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Event level deescalated successfully",
	})
}

//endregion
