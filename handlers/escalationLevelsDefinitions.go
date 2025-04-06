package handlers

import (
	"dogeplus-backend/database"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EscalationLevelsDefinitionsRequest struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func GetAllEscalationLevelsDefinitions(repos *database.Repositories) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		// Get all escalation levels definitions from db
		escalationLevelDefinitions, err := repos.EscalationLevelsDefinition.GetAll()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"result": "Retrieved all escalation levels definitions",
			"length": len(escalationLevelDefinitions),
			"data":   escalationLevelDefinitions,
		})
	}
}
