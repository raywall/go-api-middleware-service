// Package middleware provides benchmarks for the middleware functions of the pricing microservice.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raywall/go-api-middleware-service/model"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

// BenchmarkDataEnrichment benchmarks the DataEnrichment middleware.
func BenchmarkDataEnrichment(b *testing.B) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)

	// Mock external API
	var receivedCorrelationID string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCorrelationID = r.Header.Get("x-app-correlationID")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"credit_score": 750.0}`))
	}))
	defer mockServer.Close()

	// Override FetchAdditionalData to use mock server
	originalFetch := FetchAdditionalData
	FetchAdditionalData = func(ctx context.Context, customerID string, correlationID string) (map[string]interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, mockServer.URL+"?customer_id="+customerID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("x-app-correlationID", correlationID)
		resp, err := MakeExternalCall(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch credit score: %w", err)
		}
		defer resp.Body.Close()
		return map[string]interface{}{"credit_score": 750.0}, nil
	}
	defer func() { FetchAdditionalData = originalFetch }()

	router := gin.New()
	router.Use(DataEnrichment())
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Prepare request body
	reqBody, _ := json.Marshal(model.PricingRequest{
		LoanAmount: 10000.0,
		CustomerID: "CUST12345",
	})

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh request for each iteration
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
	b.StopTimer()

	// Assert response (use the last response)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-app-correlationID", "test-correlation-id-123")
	router.ServeHTTP(w, req)

	assert.Equal(b, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
	assert.Equal(b, "test-correlation-id-123", receivedCorrelationID, "Expected x-app-correlationID header test-correlation-id-123, got %s", receivedCorrelationID)
}
