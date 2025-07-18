// Package api defines the HTTP handlers for the pricing microservice.
// It provides the endpoint to process loan pricing requests and orchestrates the request flow.
package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/raywall/go-api-middleware-service/model"
	"github.com/raywall/go-api-middleware-service/service"
	"go.uber.org/zap"
)

// PricingHandler manages HTTP requests for loan pricing calculations.
type PricingHandler struct {
	engine *service.DecisionEngine
	logger *zap.Logger
}

// NewPricingHandler creates a new PricingHandler with the given decision engine.
func NewPricingHandler(engine *service.DecisionEngine) *PricingHandler {
	logger, _ := zap.NewProduction()
	return &PricingHandler{engine: engine, logger: logger}
}

// CalculateRate handles the POST /pricing endpoint.
// It retrieves the request and enriched data from the context, calls the decision engine,
// sets the pricing result, and proceeds to the next middleware for validation.
// If any step fails, it returns an appropriate error response and aborts the request.
func (h *PricingHandler) CalculateRate(c *gin.Context) {
	// Use context.Background() if c.Request is nil (for testing)
	ctx := context.Background()
	if c.Request != nil {
		ctx = c.Request.Context()
	}

	// Retrieve pricing request
	req, exists := c.Get("pricing_request")
	if !exists {
		h.logger.Error("Missing pricing request")
		c.JSON(400, gin.H{"error": "missing pricing request"})
		c.Abort()
		return
	}

	// Type assertion for pricing request
	pricingReq, ok := req.(model.PricingRequest)
	if !ok {
		h.logger.Error("Invalid pricing request type")
		c.JSON(400, gin.H{"error": "invalid pricing request type"})
		c.Abort()
		return
	}

	// Retrieve enriched data
	enrichedData, exists := c.Get("enriched_data")
	if !exists {
		h.logger.Error("Missing enriched data")
		c.JSON(400, gin.H{"error": "missing enrichment data"})
		c.Abort()
		return
	}

	// Type assertion for enriched data
	data, ok := enrichedData.(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid enriched data type")
		c.JSON(400, gin.H{"error": "invalid enriched data type"})
		c.Abort()
		return
	}

	// Calculate rate
	rate, err := h.engine.CalculateRate(ctx, pricingReq, data)
	if err != nil {
		h.logger.Error("Decision engine failed", zap.Error(err))
		c.JSON(500, gin.H{"error": "decision engine failed: " + err.Error()})
		c.Abort()
		return
	}

	// Set pricing result
	c.Set("pricing_result", rate)
	h.logger.Info("Pricing result set", zap.Float64("rate", rate))

	// Proceed to Guardrails middleware
	c.Next()

	// Check if request was aborted by middleware
	if c.IsAborted() {
		h.logger.Warn("Request aborted by middleware", zap.Any("context_keys", c.Keys))
		return
	}

	// Retrieve validated rate
	validatedRate, exists := c.Get("validated_rate")
	if !exists {
		h.logger.Error("Missing validated rate")
		c.JSON(500, gin.H{"error": "no validated rate"})
		c.Abort()
		return
	}

	// Type assertion for validated rate
	rateVal, ok := validatedRate.(float64)
	if !ok {
		h.logger.Error("Invalid validated rate type")
		c.JSON(500, gin.H{"error": "invalid validated rate type"})
		c.Abort()
		return
	}

	// Return success response
	c.JSON(200, model.PricingResponse{Rate: rateVal, Status: "success"})
}
