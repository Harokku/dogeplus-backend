package handlers

import "github.com/gofiber/fiber/v2"

// HomeHandler handles the home page route.
func HomeHandler(ctx *fiber.Ctx) error {
	return ctx.SendString("Welcome to the Home Page!")
}

// VersionLandingHandler  landing handler
func VersionLandingHandler(version string) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to " + version + " api")
	}
}
