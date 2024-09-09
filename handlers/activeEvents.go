package handlers

import (
	"dogeplus-backend/broadcast"
	"dogeplus-backend/database"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"strconv"
)

type eventRequest struct {
	Categories      []string `json:"categories"`
	EventNumber     int      `json:"event_number"`
	CentralId       string   `json:"central_id"`
	EscalationLevel string   `json:"escalation_level"`
}

type updateEventRequest struct {
	UUID       uuid.UUID `json:"uuid"`
	Status     string    `json:"status"`
	ModifiedBy string    `json:"modified_by"`
	IpAddress  string    `json:"ip_address"`
}

// CreateNewEvent is a handler function that creates a new event based on the provided categories, event number, and central ID.
// It expects a JSON request body containing the categories, event number, and central ID.
// If the body parsing fails, it returns a "400 Bad Request" error.
// If the categories field is empty, it returns a "400 Bad Request" error.
// It retrieves tasks from the repository based on the provided categories.
// If retrieving tasks fails, it returns a "500 Internal Server Error" error.
// It creates a new event using the retrieved task list, event number, and central ID.
// If creating the event fails, it returns a "500 Internal Server Error" error.
// If the request is successful, it returns a JSON response with the "Result" field set to "Events Created".
// repos is a pointer to a database.Repositories struct that contains the repositories for managing tasks and active events.
// ctx is a pointer to a fiber.Ctx object representing the HTTP request context.
func CreateNewEvent(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var body eventRequest
		err := ctx.BodyParser(&body)
		if err != nil {
			// Error while parsing body
			log.Errorf("Error parsing body: %s\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if len(body.Categories) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: Categories field should not be empty")
		}

		// Get tasks from body list
		taskList, err := repos.Tasks.GetByCategories(body.Categories)
		if err != nil {
			// Error while retrieving tasks
			log.Errorf("Error retrieving tasks: %s\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tasks")
		}

		// Filter tasks based on escalation level
		var filteredTasks []database.Task
		escalationPriority := map[string]int{
			"allarme":   1,
			"emergenza": 2,
			"incidente": 3,
		}

		for _, task := range taskList {
			if escalationPriority[task.EscalationLevel] <= escalationPriority[body.EscalationLevel] {
				filteredTasks = append(filteredTasks, task)
			}
		}

		// Create new event from taskList
		err = repos.ActiveEvents.CreateFromTaskList(filteredTasks, body.EventNumber, body.CentralId)
		if err != nil {
			// Error while creating new event
			log.Errorf("Error creating event: %s\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create event")
		}

		return ctx.JSON(fiber.Map{"Result": "Events Created"})
	}
}

// GetSingleEvent is a handler function that retrieves a single event from the database based on the provided central ID.
// It expects a JSON request body containing the central ID of the event.
// If the central ID is empty, it returns a "400 Bad Request" error.
// If no events are found, it returns a "404 Not Found" error along with the empty event and task lists.
// If multiple events are found, it returns a "300 Multiple Choices" error along with the event and task lists.
// For any other error, it returns a "500 Internal Server Error" along with the event and task lists.
// If the request is successful, it returns the event and task lists.
//
// repos is a pointer to a database.Repositories struct that contains the repositories for managing tasks and active events.
// ctx is a pointer to a fiber.Ctx object representing the HTTP request context.
func GetSingleEvent(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var body eventRequest

		body.CentralId = ctx.Params("central_id")
		//err := ctx.BodyParser(&body)
		//if err != nil {
		//	// Error while parsing body
		//	log.Errorf("Error parsing body: %s\n", err)
		//	return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		//}
		//
		if body.CentralId == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: CentralId field should not be empty")
		}

		// Get tasks for specified centralId
		taskList, events, err := repos.ActiveEvents.GetByCentralID(body.CentralId)
		if err != nil {
			switch err.(type) {
			case *database.NoEventsFoundError:
				return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"Result": "Event not found",
					"Events": events,
					"Tasks":  taskList,
				})
			case *database.MultipleEventsIdError:
				return ctx.Status(fiber.StatusMultipleChoices).JSON(fiber.Map{
					"Result": "Multiple events found",
					"Events": events,
					"Tasks":  taskList,
				})
			default:
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"Result": "Internal server error",
					"Events": events,
					"Tasks":  taskList,
				})
			}
		}

		return ctx.JSON(fiber.Map{
			"Result": "Event Found",
			"Events": events,
			"Tasks":  taskList,
		})
	}
}

// GetSpecificEvent is a handler function that retrieves a specific event based on the provided event number and central ID.
// It expects a JSON request body containing the event number and central ID.
// If the body parsing fails, it returns a "400 Bad Request" error.
// If the central ID or event number field is empty, it returns a "400 Bad Request" error.
// It retrieves the specific event from the repository based on the provided event number and central ID.
// If the event is not found, it returns a "404 Not Found" error.
// If retrieving the event fails for any other reason, it returns a "500 Internal Server Error" error.
// If the request is successful, it returns a JSON response with the "Result" field set to "Event Found" and the retrieved event.
// repos is a pointer to a database.Repositories struct that contains the repositories for managing tasks and active events.
// ctx is a pointer to a fiber.Ctx object representing the HTTP request context.
func GetSpecificEvent(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var body eventRequest

		body.CentralId = ctx.Params("central_id")
		//err := ctx.BodyParser(&body)
		//if err != nil {
		//	// Error while parsing body
		//	log.Errorf("Error parsing body: %s\n", err)
		//	return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		//}

		if body.CentralId == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: CentralId field should not be empty")
		}

		// Read event number from url param
		eventNumber, err := strconv.Atoi(ctx.Params("event_nr"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: eventNumber should be an integer")
		}

		// Check if event number is not null
		if eventNumber == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: eventNumber should not be zero")
		}

		// Get task for specified centralId and eventNumber
		taskList, err := repos.ActiveEvents.GetByCentralAndNumber(eventNumber, body.CentralId)
		if err != nil {
			switch err.(type) {
			case *database.NoEventsFoundError:
				return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"Result": "Event not found",
					"Tasks":  taskList,
				})
			default:
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"Result": "Internal server error",
					"Tasks":  taskList,
				})
			}
		}

		return ctx.JSON(fiber.Map{
			"Result": "Event Found",
			"Tasks":  taskList,
		})
	}
}

// UpdateEventTask is a handler function that updates the status of an active event in the database.
// It expects a JSON request body containing the UUID, status, and modified by fields.
// If the body parsing fails, it returns a "400 Bad Request" error.
// If the UUID field is empty, it returns a "400 Bad Request" error.
// If the status field is empty, it returns a "400 Bad Request" error.
// If the modified by field is empty, it returns a "400 Bad Request" error.
// It retrieves the client's IP address from the request context and updates the event's IP address field.
// It updates the event's status, IP address, and modified by fields in the database.
// If updating the event fails, it returns a "500 Internal Server Error" error.
// TODO: Implement broadcast to all clients
// If the request is successful, it returns a JSON response with the "Result" field set to "Event Task Updated"
// and the updated event information in the "Events" field.
// repos is a pointer to a database.Repositories struct that contains the repositories for managing active events.
// ctx is a pointer to a fiber.Ctx object representing the HTTP request context.
func UpdateEventTask(repos *database.Repositories, cm *broadcast.ConnectionManager) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var body updateEventRequest
		err := ctx.BodyParser(&body)
		if err != nil {
			// Error while parsing body
			log.Errorf("Error parsing body: %s\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		// Check if uuid is nil
		if body.UUID == uuid.Nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: UUID field should not be empty")
		}

		// Check if status is nil
		if body.Status == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: Status field should not be empty")
		}

		//check in Modified by is nil
		if body.ModifiedBy == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: ModifiedBy field should not be empty")
		}

		// Retrieve client's ip, overwrite eventual passed in value
		body.IpAddress = ctx.IP()

		// Actually update the event in db
		updatedTask, err := repos.ActiveEvents.UpdateStatus(body.UUID, body.Status, body.ModifiedBy, body.IpAddress)
		if err != nil {
			log.Errorf("Error updating event task: %s\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update event task")
		}

		// Build map for both response and broadcast
		updatedTaskMap := fiber.Map{
			"Result": "Event Task Updated",
			"Events": updatedTask,
		}

		// Send broadcast response via connection manager in JSON format
		// If error skip broadcast phase
		updatedTaskJson, err := json.Marshal(updatedTaskMap)
		if err != nil {
			log.Errorf("Error marshalling updated task: %s\n", err)
		} else {
			cm.Broadcast(updatedTaskJson)
		}

		// Send response wia http
		return ctx.JSON(updatedTaskMap)
	}
}
