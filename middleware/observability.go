// Package middleware provides middleware functions for the pricing microservice.
// This file implements the observability middleware for logging metrics.
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// StatsClient defines the interface for sending metrics to a statsd-compatible service.
type StatsClient interface {
	Incr(name string, tags []string, rate float64) error
	Histogram(name string, value float64, tags []string, rate float64) error
}

// Observability logs request metrics to a statsd-compatible service, including request count, duration, and errors.
// It captures the request's endpoint, method, status, and duration, and sends them to the provided StatsClient.
func Observability(client StatsClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Log metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		tags := []string{
			"endpoint:" + c.Request.URL.Path,
			"method:" + c.Request.Method,
			"status:" + strconv.Itoa(status), // Convert status to string properly
		}

		client.Incr("request.count", tags, 1)
		client.Histogram("request.duration", duration, tags, 1)

		if c.Errors != nil {
			client.Incr("request.error", tags, 1)
		}
	}
}
