// Package main serves as the entry point for the record query microservice.
// It handles the configuration of the HTTP server, middleware setup, and service initialization.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/gin-gonic/gin"
	"github.com/raywall/go-middleware"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	_ "github.com/go-sql-driver/mysql"
)

// Global variables for logging, routing, and middleware chain.
var (
	// logger is the structured logger for application logging.
	logger *slog.Logger
	// engine is the Gin router for handling HTTP requests.
	engine *gin.Engine
	// chain is the middleware chain for processing requests.
	chain *middleware.Chain
)

// init initializes the microservice by setting up the Datadog client, Gin router, and middleware chain.
func init() {
	// Start the Datadog tracer for distributed tracing.
	tracer.Start()
	// Ensure the tracer is stopped when the application exits.
	defer tracer.Stop()

	// Initialize the Datadog statsd client for observability metrics.
	// The client connects to the Datadog agent at 127.0.0.1:8125 with the namespace "sample_service".
	statsdClient, _ = statsd.New("127.0.0.1:8125", statsd.WithNamespace("sample_service"))
	// Ensure the statsd client is closed when the application exits.
	defer statsdClient.Close()

	// Initialize the structured logger to output logs to stdout.
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Configure the Gin router with default middleware (logging and recovery).
	engine = gin.Default()

	// Create a middleware chain with observability and business logic middleware.
	// Additional middleware (e.g., validation, authentication) can be added here.
	chain = middleware.NewChain(
		middleware.Observability(logger),
		// Example: middleware.Validation(), middleware.Auth(),
		businessLogic(logger),
	)
}

// main starts the HTTP server and sets up the routes.
func main() {
	// Establish a connection to the MySQL database.
	if err := conectar(); err != nil {
		// Log the error and exit if the connection fails.
		logger.Error("Erro ao conectar ao MySQL", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Define the POST /consulta route for querying data.
	engine.POST("/consulta", func(c *gin.Context) {
		// Record a metric for the consulta action.
		_ = registrarMetrica([]string{"action:consulta"})

		// Bind the JSON payload to the Payload struct.
		var payload Payload
		if err := c.ShouldBindJSON(&payload); err != nil {
			// Return a 400 Bad Request response if the payload is invalid.
			c.JSON(400, gin.H{"error": "Payload inválido: " + err.Error()})
			return
		}

		// Extract the request context.
		ctx := c.Request.Context()
		// Process the request through the middleware chain.
		_, result, err := chain.Then(ctx, payload)
		if err != nil {
			// Log the error and return a 500 Internal Server Error response.
			logger.Error("Erro na execução", slog.String("err", err.Error()))
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Log the successful result and return a 200 OK response with the result.
		logger.Info("Resultado final", slog.Any("output", result))
		c.JSON(200, result)
	})

	// Start the HTTP server on port 8080.
	_ = engine.Run(":8080")
}

// businessLogic defines the core logic for processing requests.
// It takes a logger and returns a middleware function that processes the payload and queries the database.
func businessLogic(logger *slog.Logger) middleware.MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		// Assert the input is a Payload struct.
		payload, ok := input.(Payload)
		if !ok {
			// Log and return an error if the payload is invalid.
			logger.Error("Payload inválido")
			return ctx, nil, fmt.Errorf("payload inválido")
		}

		// Query the database with the provided UserID.
		data, err := consultar(payload.UserID)
		if err != nil {
			// Log and return an error if the query fails.
			logger.Error("Erro ao consultar dados", slog.String("err", err.Error()))
			return ctx, nil, fmt.Errorf("erro ao consultar dados: %w", err)
		}

		// Return the context, query results, and no error.
		return ctx, *data, nil
	}
}
