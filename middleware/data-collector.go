// Package middleware provides middleware functions for the pricing microservice.
// This file implements the data enrichment middleware.
package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raywall/go-api-middleware-service/model"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// DataEnrichment enriches the pricing request with additional data (e.g., credit score).
// It binds the incoming JSON request, extracts the x-app-correlationID header,
// fetches enrichment data, and stores it in the Gin context.
// If any step fails, it aborts the request with an error response.
func DataEnrichment() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()

		// Extract x-app-correlationID header
		correlationID := c.GetHeader("x-app-correlationID")
		if correlationID == "" {
			logger.Warn("Missing x-app-correlationID header")
			correlationID = "unknown"
		} else {
			logger.Info("Extracted x-app-correlationID", zap.String("correlation_id", correlationID))
		}
		c.Set("correlation_id", correlationID)

		var req model.PricingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind JSON",
				zap.Error(err),
				zap.Any("request_body", c.Request.Body),
				zap.String("correlation_id", correlationID))
			c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
			c.Abort()
			return
		}

		// Validate request
		if req.CustomerID == "" || req.LoanAmount <= 0 {
			logger.Error("Invalid request data",
				zap.String("customer_id", req.CustomerID),
				zap.Float64("loan_amount", req.LoanAmount),
				zap.String("correlation_id", correlationID))
			c.JSON(500, gin.H{"error": "failed to enrich data: invalid customer ID or loan amount"})
			c.Abort()
			return
		}

		// Fetch additional data
		enrichedData, err := FetchAdditionalData(c.Request.Context(), req.CustomerID, correlationID)
		if err != nil {
			logger.Error("Failed to fetch additional data",
				zap.Error(err),
				zap.String("customer_id", req.CustomerID),
				zap.String("correlation_id", correlationID))
			c.JSON(500, gin.H{"error": "failed to enrich data: " + err.Error()})
			c.Abort()
			return
		}

		// Ensure enrichedData is not nil
		if enrichedData == nil {
			enrichedData = map[string]interface{}{}
		}

		// Store enriched data in Gin context
		c.Set("enriched_data", enrichedData)
		c.Set("pricing_request", req)
		logger.Info("Data enriched successfully",
			zap.Any("enriched_data", enrichedData),
			zap.Any("pricing_request", req),
			zap.String("correlation_id", correlationID))

		c.Next()
	}
}

// FetchAdditionalData is a function variable that retrieves additional data for the given customer ID.
// It simulates an external API call (e.g., to fetch a credit score) and includes
// the x-app-correlationID header for traceability.
var FetchAdditionalData = func(ctx context.Context, customerID string, correlationID string) (map[string]interface{}, error) {
	if customerID == "" {
		return nil, fmt.Errorf("invalid customer ID")
	}

	// Simulate external API call with correlation ID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/credit-score?customer_id="+customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("x-app-correlationID", correlationID)

	// Use utility function to make the HTTP call
	resp, err := MakeExternalCall(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credit score: %w", err)
	}
	defer resp.Body.Close()

	// Simulate response parsing (replace with actual API response handling)
	return map[string]interface{}{"credit_score": 750.0}, nil
}

// MakeExternalCall executes an HTTP request with the x-app-correlationID header.
// It is a reusable utility for making external API calls with traceability.
func MakeExternalCall(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned status %d", resp.StatusCode)
	}
	return resp, nil
}
