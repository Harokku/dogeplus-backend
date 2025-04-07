package handlers

import (
	"dogeplus-backend/config"
	"dogeplus-backend/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"slices"
)

type taskCompletionInfo struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// -------------------------
// Task completion info functions
// -------------------------

// region TaskCompletionInfo

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
	EventNumber   int            `json:"eventNumber"`
	NewLevel      database.Level `json:"newLevel"`
	Direction     string         `json:"direction"`
	IncidentLevel string         `json:"incidentLevel"`
}

// PostNewOverview handles the posting of new overview records to the database.
// It parses the request body, validates it, and uses the repository to add the overview.
func PostNewOverview(repos *database.Repositories) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Parse request body
		var request database.Overview
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse request",
			})
		}

		// Add the overview
		err := repos.Overview.Add(request)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Sync escalation levels in memory map
		repos.EscalationLevelsAggregation.Add(request.EventNumber, database.Level(request.Level))

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Overview added successfully",
		})
	}
}

// GetAllEscalationLevels handles the HTTP request to retrieve all escalation levels.
// It fetches the escalation levels from the database and returns them as a JSON response.
func GetAllEscalationLevels(c *fiber.Ctx) error {
	escalationLevels := database.GetEscalationLevelsInstance(nil)
	levelData := escalationLevels.GetLevels()
	return c.JSON(levelData)
}

// PostEscalate handles HTTP POST requests to escalate an event's level.
// It reads the request body to get the eventNumber and newLevel, escalates the event level,
// and updates the overview. It returns a JSON response indicating success or any error.
func PostEscalate(repos *database.Repositories, confg config.Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Parse request body
		var request EscalateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse request",
			})
		}

		// Get escalation map instance
		escalationLevels := database.GetEscalationLevelsInstance(nil)

		// Get actual escalation levels
		actualLevels := escalationLevels.GetLevels()
		oldLevel := actualLevels[request.EventNumber]

		// Call the Escalate method
		err := escalationLevels.Escalate(request.EventNumber, request.NewLevel)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Sync overview with new escalation level
		err = repos.Overview.UpdateLevelByEventNumber(request.EventNumber, request.NewLevel)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Get actual event overview snapshot
		actualOverview, err := repos.Overview.GetOverviewById(request.EventNumber)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// FIXME: Implement correct filtering using db query
		// Get new level tasks
		// newTasks, err := repos.Tasks.GetGyCategoryAndEscalationLevel(actualOverview.Type, string(oldLevel), string(request.NewLevel), request.IncidentLevel)
		newTasks, err := repos.Tasks.GetByCategories(actualOverview.Type)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		newFilteredTasks, err := database.FilterTasksForEscalation(newTasks, actualOverview.Type, string(oldLevel), string(request.NewLevel), request.IncidentLevel)

		// Get local Tasks based on selection
		var tasksToUse []database.Task
		isMergedTasks := false
		// Load the correct local task file
		f, err := config.LoadExcelFile(confg, actualOverview.CentralId)
		if err == nil {

			// Parse the file
			localTasks, err := database.ParseXLSXToTasks(f)
			if err != nil {
				log.Errorf("Error parsing local task file: %s\n", err)
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse local task file")
			}

			// Filter the local file based on request body parameters
			filteredLocalTasks, err := database.FilterTasksForEscalation(localTasks, actualOverview.Type, string(oldLevel), string(request.NewLevel), request.IncidentLevel)
			if err != nil {
				log.Errorf("Error filtering local tasks for escalation: %s\n", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to filter local tasks for escalation",
				})
			}

			// Merge task lists
			tasksToUse, err = database.MergeTasks(filteredLocalTasks, newFilteredTasks)
			if err != nil {
				// Error while merging tasks
				log.Errorf("Error merging tasks: %s\n", err)
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to merge tasks")
			}
			// Set isMergedTasks to true to signal that the data need to be sorted
			isMergedTasks = true
		} else {
			tasksToUse = newFilteredTasks
		}

		// Sort merged tasks by priority if merging occurred
		if isMergedTasks {
			slices.SortStableFunc(tasksToUse, func(a, b database.Task) int {
				if a.Priority < b.Priority {
					return -1
				}
				if a.Priority > b.Priority {
					return 1
				}
				return 0
			})
		}

		// Add new tasks to active events
		err = repos.ActiveEvents.CreateFromTaskList(tasksToUse, actualOverview.EventNumber, actualOverview.CentralId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Keep in memory completion metrics cache in sync
		taskCompletionInstance := database.GetTaskCompletionMapInstance(nil)
		taskCompletionInstance.AddMultipleNotDoneTasks(request.EventNumber, len(tasksToUse))

		// build a map for realtime update
		//updatedEscalation := fiber.Map{
		//	"Result":      "Escalation level updated",
		//	"EventNumber": request.EventNumber,
		//	"AddedTasks":  len(newTasks),
		//}
		//
		//// Send broadcast response via connection manager in JSON format
		//// If error skip broadcast phase
		//updatedEscalationJson, err := json.Marshal(updatedEscalation)
		//if err != nil {
		//	log.Errorf("Failed to marshal updated escalation level to JSON: %v\n", err)
		//} else {
		//	cm.Broadcast(updatedEscalationJson)
		//}

		// Return success response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Event level escalated successfully",
		})
	}
}

// PostDeEscalate handles de-escalation of an event level based on the provided request data and updates the escalation map.
func PostDeEscalate(repos *database.Repositories) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
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
		// Sync overview with new escalation level
		err = repos.Overview.UpdateLevelByEventNumber(request.EventNumber, request.NewLevel)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return success response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Event level deescalated successfully",
		})
	}
}

// GetAllEscalationDetails retrieves all escalation details from the database and returns them in the HTTP response.
func GetAllEscalationDetails(repos *database.Repositories) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Get overview from db
		overview, err := repos.Overview.GetAllOverview()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  err.Error(),
				"detail": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(
			fiber.Map{
				"result": "Retrieved all monitored events",
				"length": len(overview),
				"data":   overview,
			},
		)
	}
}

// GetEscalationDetailsByCentralId fetches the overview details for a given central ID or all overviews if central_id is "GLOBAL".
func GetEscalationDetailsByCentralId(repos *database.Repositories) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		centralID := c.Params("central_id")
		var overview []database.Overview
		var err error

		// Check if central_id equals "GLOBALE"
		if centralID == "GLOBALE" {
			overview, err = repos.Overview.GetAllOverview()
		} else {
			overview, err = repos.Overview.GetOverviewByCentralId(centralID)
		}

		// Return error response if database operation fails
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  err.Error(),
				"detail": err.Error(),
			})
		}

		// Return the response as JSON
		return c.Status(fiber.StatusOK).JSON(
			fiber.Map{
				"result": "Retrieved all monitored events",
				"length": len(overview),
				"data":   overview,
			},
		)
	}
}

// GetEscalationDetailsByCentralIdAndEventNumber fetches the overview details for a given central ID and event number.
func GetEscalationDetailsByCentralIdAndEventNumber(repos *database.Repositories) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		centralID := c.Params("central_id")
		eventNumber, err := c.ParamsInt("event_number")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Invalid event number",
				"detail": err.Error(),
			})
		}

		overview, err := repos.Overview.GetOverviewByCentralIdAndEventNumber(centralID, eventNumber)

		// Return error response if database operation fails
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  err.Error(),
				"detail": err.Error(),
			})
		}

		// Return the response as JSON
		return c.Status(fiber.StatusOK).JSON(
			fiber.Map{
				"result": "Retrieved events for central ID and event number",
				"length": len(overview),
				"data":   overview,
			},
		)
	}
}

//endregion
