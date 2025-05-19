package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sol1corejz/enricher/internal/handlers"
	"github.com/sol1corejz/enricher/internal/services/enricher"
	"github.com/sol1corejz/enricher/internal/storage/postgres"
	"log/slog"
)

type App struct {
	FiberSrv *fiber.App
}

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

	fiberApp.Get("/", handlers.DataWithFilters)
	fiberApp.Post("/add", handlers.Add)
	fiberApp.Post("/delete", handlers.Delete)
	fiberApp.Post("/edit", handlers.Edit)

	return &App{
		FiberSrv: fiberApp,
	}
}
