// Package middleware provides middleware functions for the pricing microservice.
// This file implements the guardrails middleware for validating pricing results.
package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Guardrails validates the pricing result to prevent invalid or erroneous rates.
// It checks if the pricing result exists and falls within acceptable bounds (0 to 20).
// If the request is already aborted, it skips validation to avoid overwriting the response.
func Guardrails() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, _ := zap.NewProduction()
		defer logger.Sync()

		// Skip if request is already aborted
		if c.IsAborted() {
			logger.Warn("Skipping Guardrails for aborted request")
			return
		}

		// Check for pricing result
		result, exists := c.Get("pricing_result")
		if !exists {
			logger.Error("No pricing result found")
			c.JSON(500, gin.H{"error": "no pricing result"})
			c.Abort()
			return
		}

		// Validate rate
		rate, ok := result.(float64)
		if !ok {
			logger.Error("Invalid pricing result type")
			c.JSON(500, gin.H{"error": "invalid pricing result type"})
			c.Abort()
			return
		}

		if rate < 0 || rate > 20 {
			logger.Error("Invalid rate", zap.Float64("rate", rate))
			c.JSON(500, gin.H{"error": "invalid rate"})
			c.Abort()
			return
		}

		// Set validated rate
		logger.Info("Validated rate", zap.Float64("rate", rate))
		c.Set("validated_rate", rate)
		c.Next()
	}
}
