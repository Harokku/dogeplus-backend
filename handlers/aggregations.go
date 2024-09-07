package handlers

import (
	"dogeplus-backend/database"
	"github.com/gofiber/fiber/v2"
)

type taskCompletionInfo struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

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

func GetAllEscalationLevels(c *fiber.Ctx) error {
	escalationLevels := database.GetEscalationLevelsInstance(nil)
	levelData := escalationLevels.GetLevels()
	return c.JSON(levelData)
}
