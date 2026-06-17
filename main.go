package main

import (
	"fmt"
	"log"

	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	// Initialize default config
	app := fiber.New()
	app.Use(logger.New())

	// Route de vérification de la santé et metrics
	app.Get("/health", healthcheck.New())
	app.Get("/metrics", monitor.New(monitor.Config{APIOnly: true}))
	// Route de base de l'API
	app.Get("/api", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Welcome to the open ITER dpp API",
			"version": "1.0.0",
		})
	})

	// Démarrer le serveur sur le port 7000
	port := 7000
	log.Printf("Starting server on port %d...", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
