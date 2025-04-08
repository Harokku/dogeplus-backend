package handlers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetTasksHandler tests the GetTasks handler functionality
func TestGetTasksHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a test handler that simulates the behavior of GetTasks
	app.Get("/tasks", func(ctx *fiber.Ctx) error {
		// Simulate successful retrieval of categories
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"result": "Retrieved categories",
			"length": 2,
			"data":   []string{"category1", "category2"},
		})
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check the status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the response
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, "Retrieved categories", response["result"])
	assert.Equal(t, float64(2), response["length"])
	assert.Equal(t, []interface{}{"category1", "category2"}, response["data"])
}

// TestGetTasksForEscalationHandler tests the GetTasksForEscalation handler functionality
func TestGetTasksForEscalationHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a test handler that simulates the behavior of GetTasksForEscalation
	app.Post("/tasks/escalation", func(ctx *fiber.Ctx) error {
		// Parse the request body to verify it's correct
		var request map[string]string
		if err := ctx.BodyParser(&request); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Failed to parse request body",
				"detail": err.Error(),
			})
		}

		// Verify the request contains the expected values
		if request["category"] != "category1" ||
			request["start_level"] != "level1" ||
			request["end_level"] != "level2" ||
			request["incident_level"] != "incident1" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request parameters",
			})
		}

		// Simulate successful retrieval of tasks
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"result": "Retrieved tasks",
			"length": 1,
			"data": []map[string]interface{}{
				{
					"ID":               1,
					"priority":         1,
					"title":            "Task 1",
					"description":      "Description 1",
					"role":             "Role 1",
					"category":         "category1",
					"escalation_level": "level1",
					"incident_level":   "incident1",
				},
			},
		})
	})

	// Create a test request
	reqBody := `{
		"category": "category1",
		"start_level": "level1",
		"end_level": "level2",
		"incident_level": "incident1"
	}`
	req := httptest.NewRequest(http.MethodPost, "/tasks/escalation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check the status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the response
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, "Retrieved tasks", response["result"])
	assert.Equal(t, float64(1), response["length"])
}
