// Package main is the entry point for the pricing microservice.
// It sets up the HTTP server, configures middleware, and starts the service.
package main

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/gin-gonic/gin"
	"github.com/raywall/go-api-middleware-service/api"
	"github.com/raywall/go-api-middleware-service/middleware"
	"github.com/raywall/go-api-middleware-service/service"
)

// main initializes and starts the pricing microservice.
// It configures the Datadog client, decision engine, and HTTP routes with middleware.
func main() {
	// Initialize Datadog client for observability
	statsdClient, _ := statsd.New("127.0.0.1:8125", statsd.WithNamespace("pricing_service"))

	// Initialize decision engine service
	decisionEngine := service.NewDecisionEngine()
	pricingHandler := api.NewPricingHandler(decisionEngine)

	// Set up Gin router
	r := gin.Default()

	// Configure POST /pricing endpoint with middleware chain
	r.POST("/pricing",
		middleware.Observability(statsdClient),
		middleware.DataEnrichment(),
		pricingHandler.CalculateRate,
		middleware.Guardrails(),
		middleware.Observability(statsdClient),
	)

	// Start the HTTP server on port 8080
	r.Run(":8080")
}
