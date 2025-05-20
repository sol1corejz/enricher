package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/sol1corejz/enricher/docs"
	"github.com/sol1corejz/enricher/internal/handlers"
	"github.com/sol1corejz/enricher/internal/services/enricher"
	"github.com/sol1corejz/enricher/internal/storage/postgres"
	"log/slog"
)

type App struct {
	FiberSrv *fiber.App
}

// @title User Enricher API
// @version 1.0
// @description API for managing and enriching user data
// @host localhost:8080
// @BasePath /
func New(log *slog.Logger) *App {

	storage, err := postgres.New()
	log.Info("connected to database")
	if err != nil {
		panic(err)
	}

	enricherService := enricher.New(log, storage)

	fiberApp := fiber.New()
	fiberApp.Use(func(c *fiber.Ctx) error {
		c.Locals("enricherService", enricherService)
		return c.Next()
	})

	// Swagger UI
	fiberApp.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: false,
	}))

	fiberApp.Get("/", handlers.DataWithFilters)
	fiberApp.Post("/add", handlers.Add)
	fiberApp.Post("/delete", handlers.Delete)
	fiberApp.Post("/edit", handlers.Edit)

	return &App{
		FiberSrv: fiberApp,
	}
}
