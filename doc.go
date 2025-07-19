// Package main implements a record query microservice that provides an HTTP API
// for querying records from a MySQL database. It includes middleware for
// observability, request validation, and business logic processing, along with
// metrics collection using Datadog.
//
// # Overview
//
// The microservice is designed to handle POST requests to the /consulta endpoint,
// allowing clients to query records by user ID. It uses the Gin framework for
// routing, Datadog for tracing and metrics, and a MySQL database for data storage.
// The middleware chain ensures observability and extensibility for additional
// features like authentication or validation.
//
// # Key Components
//
//   - main.go: Configures the HTTP server, initializes middleware, and defines the
//     /consulta endpoint.
//   - database.go: Handles MySQL database connections and query operations.
//   - metrics.go: Manages metrics collection using the Datadog statsd client.
//   - models.go: Defines the data structures (Registro, Registros, Payload) used for
//     database queries and API interactions.
//
// # Usage Example
//
// To query records, send a POST request to the /consulta endpoint with a JSON
// payload containing the user_id field. The response will include the matching
// records or an error message.
//
// Example Request:
//
//	POST /consulta HTTP/1.1
//	Content-Type: application/json
//
//	{
//	  "user_id": "12345",
//	  "name": "optional-name"
//	}
//
// Example Response (Success):
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//
//	[
//	  {
//	    "id": "12345",
//	    "name": "John Doe"
//	  }
//	]
//
// Example Response (Error):
//
//	HTTP/1.1 400 Bad Request
//	Content-Type: application/json
//
//	{
//	  "error": "Payload inválido: missing user_id"
//	}
//
// # Configuration
//
// The microservice connects to a MySQL database using the connection string
// "dbuser:senha@tcp(localhost:3306)/db?parseTime=true". Ensure the database is
// running and accessible. Datadog tracing and metrics require a Datadog agent
// running at 127.0.0.1:8125. The server listens on port 8080 by default.
//
// # Error Handling
//
// The microservice handles errors gracefully, logging them using the slog package
// and returning appropriate HTTP status codes (e.g., 400 for invalid payloads, 500
// for server errors). Metrics are recorded for each request to the /consulta
// endpoint using the Datadog statsd client.
package main
