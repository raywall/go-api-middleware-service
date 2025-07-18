// Package main provides a microservice for calculating loan pricing rates.
// It acts as an intermediary between an API gateway and a decision engine,
// applying middleware layers for observability, data enrichment, and guardrails
// to ensure resilience, performance, scalability, and observability for handling
// millions of requests.
//
// The service exposes a single HTTP POST endpoint (/pricing) that processes
// loan pricing requests, enriches them with additional data, calls an external
// decision engine to calculate rates, validates the results, and logs metrics
// to Datadog for monitoring.
//
// Key components:
//   - Middleware: Handles observability (Datadog metrics), data enrichment, and guardrails.
//   - Service: Integrates with an external decision engine using a circuit breaker for resilience.
//   - API: Defines the HTTP endpoint and orchestrates the request flow.
//
// Usage:
//   - Configure the service with environment variables or a configuration file.
//   - Run the service using `go run main.go`.
//   - Test the endpoint with the provided `test_pricing.sh` (Linux/macOS) or `test_pricing.bat` (Windows).
package main
