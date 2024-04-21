package router

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:           "DogePlus Backend",
		BodyLimit:         1 * 1024 * 1024,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		EnablePrintRoutes: true,
	})
	return app
}
