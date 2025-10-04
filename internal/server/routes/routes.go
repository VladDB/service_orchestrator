package routes

import (
	"service_orchestrator/internal/server/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/processes", handlers.GetAllProcesses)
	api.Get("/processes/:id", handlers.GetProcessById)
	api.Post("/processes", handlers.AllProcessesAction)
	api.Post("/processes/:id", handlers.ProcessActionById)
}
