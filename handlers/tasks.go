package handlers

import (
	"bytes"
	"dogeplus-backend/database"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"io"
	"path/filepath"
)

type EscalationRequest struct {
	Category      string `json:"category"`
	StartLevel    string `json:"start_level"`
	EndLevel      string `json:"end_level"`
	IncidentLevel string `json:"incident_level"`
}

// UploadMainTasksFile handles the upload of an .xlsx file containing main tasks, resets the tasks table, and adds new tasks.
func UploadMainTasksFile(repos *database.Repositories) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		// Access the file:
		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Could not access file: %v", err))
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Could not open file: %v", err))
		}
		defer file.Close()

		// Ensure the uploaded file is an xlsx file
		if filepath.Ext(fileHeader.Filename) != ".xlsx" {
			return ctx.Status(fiber.StatusBadRequest).SendString("Invalid file type: only .xlsx files are accepted")
		}

		// Read file into memory
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to read file: %v", err))
		}

		// Open the file with excelize
		excelFile, err := excelize.OpenReader(bytes.NewReader(fileBytes))
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to open xlsx file: %v", err))
		}

		// Parse the xlsx file into tasks
		tasks, err := database.ParseXLSXToTasks(excelFile)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to parse xlsx file to tasks: %v", err))
		}

		// Use a transaction to ensure atomicity
		if err := repos.Tasks.WithTransaction(func(tx *database.TaskRepositoryTransaction) error {
			// Drop the existing tasks table
			if err := tx.DropTasksTable(); err != nil {
				return fmt.Errorf("failed to drop tasks table: %v", err)
			}

			// Add tasks to database
			if err := tx.BulkAdd(tasks); err != nil {
				return fmt.Errorf("failed to add tasks to database: %v", err)
			}

			return nil
		}); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		// Respond to the request
		return ctx.SendString("File processed and tasks table reset successfully.")
	}
}

// UploadLocalTasksFile handles the upload of a local tasks file via multipart form and saves it to a specified directory.
func UploadLocalTasksFile() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		// Retrieve file from the multipart form
		file, err := ctx.FormFile("file")
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).SendString("Failed to retrieve file")
		}

		// Get the filename and the extension
		filename := file.Filename

		// FIXME: Implement robust path handling reading from .env
		// Define the path to save the file
		savePath := filepath.Join("/Users/simonecrenna/GolandProjects/dogeplus-backend/db/", filename) // Saves the file in the current directory

		// Save the file to the defined path
		if err := ctx.SaveFile(file, savePath); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to save file")
		}

		// Return success response
		return ctx.SendString(fmt.Sprintf("File %s uploaded successfully", filename))
	}
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
		categoriesList, err := repos.Tasks.GetGyCategoryAndEscalationLevel(escalationRequest.Category, escalationRequest.StartLevel, escalationRequest.EndLevel, escalationRequest.IncidentLevel)
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
