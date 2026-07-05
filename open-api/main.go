package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Nombre total de requêtes HTTP traitées par l'API.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Durée des requêtes HTTP en secondes.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func metricsMiddleware(c fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	path := c.Route().Path
	method := c.Method()
	httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(c.Response().StatusCode())).Inc()

	return err
}

func setupApp() *fiber.App {
	app := fiber.New()
	app.Use(logger.New())
	app.Use(metricsMiddleware)

	// Route de vérification de la santé
	app.Get("/health", healthcheck.New())
	// Route de métriques au format Prometheus (scrapée par Grafana Alloy)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	// Route de base de l'API
	app.Get("/api", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Welcome to the open ITER dpp API",
			"version": "1.0.0",
		})
	})

	return app
}

func main() {
	app := setupApp()

	// Démarrer le serveur sur le port 7000
	port := 7000
	log.Printf("Starting server on port %d...", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
