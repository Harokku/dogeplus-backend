package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func CreateNewEvent(ctx *fiber.Ctx) error {
	var body interface{}
	err := ctx.BodyParser(&body)
	if err != nil {
		// Error while parsing body
		log.Errorf("Error parsing body: %s\n", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	bodySlice, ok := body.([]interface{})
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).SendString("Invalid request body: Body should be a JSON-encoded array")
	}

	stringSlice := []string{}
	for _, item := range bodySlice {
		str, ok := item.(string)
		if !ok {
			return ctx.Status(fiber.StatusBadRequest).SendString("Invalid request body: Body should contain only strings")
		}
		stringSlice = append(stringSlice, str)
	}
	// TODO: implement event creation logic
}
