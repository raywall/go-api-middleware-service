// Package api provides tests for the HTTP handlers of the pricing microservice.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raywall/go-api-middleware-service/middleware"
	"github.com/raywall/go-api-middleware-service/model"
	"github.com/raywall/go-api-middleware-service/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCalculateRate tests the CalculateRate handler for various scenarios.
func TestCalculateRate(t *testing.T) {
	// Initialize test dependencies
	engine := service.NewDecisionEngine()
	handler := NewPricingHandler(engine)

	// Set up Gin router
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		// Mock external API
		var receivedCorrelationID string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedCorrelationID = r.Header.Get("x-app-correlationID")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"credit_score": 750.0}`))
		}))
		defer mockServer.Close()

		// Override FetchAdditionalData to use mock server
		originalFetch := middleware.FetchAdditionalData
		middleware.FetchAdditionalData = func(ctx context.Context, customerID string, correlationID string) (map[string]interface{}, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, mockServer.URL+"?customer_id="+customerID, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("x-app-correlationID", correlationID)
			resp, err := middleware.MakeExternalCall(req)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch credit score: %w", err)
			}
			defer resp.Body.Close()
			return map[string]interface{}{"credit_score": 750.0}, nil
		}
		defer func() { middleware.FetchAdditionalData = originalFetch }()

		var capturedContext *gin.Context
		router := gin.New()
		router.POST("/pricing",
			middleware.DataEnrichment(),
			handler.CalculateRate,
			middleware.Guardrails(),
			func(c *gin.Context) {
				capturedContext = c // Set before c.Next()
				c.Next()
			},
		)

		// Create request
		reqBody, _ := json.Marshal(model.PricingRequest{
			LoanAmount: 10000.0,
			CustomerID: "CUST12345",
		})
		req := httptest.NewRequest(http.MethodPost, "/pricing", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		// Log request body for debugging
		body, _ := io.ReadAll(bytes.NewBuffer(reqBody))
		t.Logf("Request body: %s", string(body))

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
		var resp model.PricingResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "Failed to unmarshal response: %v", err)
		assert.Equal(t, 5.0, resp.Rate, "Expected rate 5.0, got %f", resp.Rate)
		assert.Equal(t, "success", resp.Status, "Expected status 'success', got %s", resp.Status)
		assert.NotNil(t, capturedContext, "Expected captured context")
		if w.Code == http.StatusOK {
			validatedRate, exists := capturedContext.Get("validated_rate")
			assert.True(t, exists, "Expected validated_rate in context")
			assert.Equal(t, 5.0, validatedRate, "Expected validated_rate 5.0, got %v", validatedRate)
			correlationID, exists := capturedContext.Get("correlation_id")
			assert.True(t, exists, "Expected correlation_id in context")
			assert.Equal(t, "test-correlation-id-123", correlationID, "Expected correlation_id test-correlation-id-123, got %v", correlationID)
			assert.Equal(t, "test-correlation-id-123", receivedCorrelationID, "Expected x-app-correlationID header test-correlation-id-123, got %s", receivedCorrelationID)
		}
	})

	t.Run("MissingPricingRequest", func(t *testing.T) {
		router := gin.New()
		router.POST("/pricing", handler.CalculateRate)

		req := httptest.NewRequest(http.MethodPost, "/pricing", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400, got %d: %s", w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "missing pricing request")
	})

	t.Run("MissingEnrichedData", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// Simulate DataEnrichment setting only pricing_request
			c.Set("pricing_request", model.PricingRequest{
				LoanAmount: 10000.0,
				CustomerID: "CUST12345",
			})
			c.Next()
		})
		router.POST("/pricing", handler.CalculateRate)

		req := httptest.NewRequest(http.MethodPost, "/pricing", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400, got %d: %s", w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "missing enrichment data")
	})
}
