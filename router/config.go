package router

import "github.com/gofiber/fiber/v2"

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:           "DogePlus Backend",
		EnablePrintRoutes: true,
	})
	return app
}
