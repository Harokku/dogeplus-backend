package handlers

import (
	"dogeplus-backend/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type eventRequest struct {
	Categories  []string `json:"categories"`
	EventNumber int      `json:"event_number"`
	CentralId   string   `json:"central_id"`
}

type updateEventRequest struct {
	UUID       uuid.UUID `json:"uuid"`
	Status     string    `json:"status"`
	ModifiedBy string    `json:"modified_by"`
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
// Example usage:
//
//	repos := &database.Repositories{...}
//	app.Post("/events", CreateNewEvent(repos))
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

		// Create new event from taskList
		err = repos.ActiveEvents.CreateFromTaskList(taskList, body.EventNumber, body.CentralId)
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
//
// Example usage:
//
//	repos := &database.Repositories{...}
//	app.Get("/event", GetSingleEvent(repos))
func GetSingleEvent(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var body eventRequest
		err := ctx.BodyParser(&body)
		if err != nil {
			// Error while parsing body
			log.Errorf("Error parsing body: %s\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

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

func UpdateEventTask(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var body updateEventRequest
		err := ctx.BodyParser(&body)
		if err != nil {
			// Error while parsing body
			log.Errorf("Error parsing body: %s\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if body.UUID == uuid.Nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: UUID field should not be empty")
		}

		if body.Status == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: Status field should not be empty")
		}

		if body.ModifiedBy == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request: ModifiedBy field should not be empty")
		}

		updatedTask, err := repos.ActiveEvents.UpdateStatus(body.UUID, body.Status, body.ModifiedBy)
		if err != nil {
			log.Errorf("Error updating event task: %s\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to update event task")
		}

		return ctx.JSON(fiber.Map{
			"Result": "Event Task Updated",
			"Events": updatedTask,
		})
	}
}
