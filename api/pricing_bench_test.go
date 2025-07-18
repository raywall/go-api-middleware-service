// Package api provides benchmarks for the HTTP handlers of the pricing microservice.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raywall/go-api-middleware-service/middleware"
	"github.com/raywall/go-api-middleware-service/model"
	"github.com/raywall/go-api-middleware-service/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// BenchmarkPricingEndpoint benchmarks the entire pricing endpoint with middleware.
func BenchmarkPricingEndpoint(b *testing.B) {
	// Initialize test dependencies
	engine := service.NewDecisionEngine()
	handler := NewPricingHandler(engine)

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

	router := gin.New()
	router.POST("/pricing",
		middleware.DataEnrichment(),
		handler.CalculateRate,
		middleware.Guardrails(),
	)

	// Prepare request body
	reqBody, _ := json.Marshal(model.PricingRequest{
		LoanAmount: 10000.0,
		CustomerID: "CUST12345",
	})

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh request for each iteration
		req := httptest.NewRequest(http.MethodPost, "/pricing", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
	b.StopTimer()

	// Assert response (use the last response)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/pricing", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-app-correlationID", "test-correlation-id-123")
	router.ServeHTTP(w, req)

	assert.Equal(b, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
	var resp model.PricingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(b, err, "Failed to unmarshal response: %v", err)
	assert.Equal(b, 5.0, resp.Rate, "Expected rate 5.0, got %f", resp.Rate)
	assert.Equal(b, "success", resp.Status, "Expected status 'success', got %s", resp.Status)
	assert.Equal(b, "test-correlation-id-123", receivedCorrelationID, "Expected x-app-correlationID header test-correlation-id-123, got %s", receivedCorrelationID)
}

// BenchmarkCalculateRate benchmarks the CalculateRate handler.
func BenchmarkCalculateRate(b *testing.B) {
	// Initialize test dependencies
	engine := service.NewDecisionEngine()
	handler := NewPricingHandler(engine)

	// Set up Gin router
	gin.SetMode(gin.TestMode)

	// Mock external API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	router := gin.New()
	router.POST("/pricing",
		middleware.DataEnrichment(),
		handler.CalculateRate,
		middleware.Guardrails(),
	)

	// Prepare request body
	reqBody, _ := json.Marshal(model.PricingRequest{
		LoanAmount: 10000.0,
		CustomerID: "CUST12345",
	})

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh request for each iteration
		req := httptest.NewRequest(http.MethodPost, "/pricing", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
	b.StopTimer()

	// Assert response (use the last response)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/pricing", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-app-correlationID", "test-correlation-id-123")
	router.ServeHTTP(w, req)

	assert.Equal(b, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
	var resp model.PricingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(b, err, "Failed to unmarshal response: %v", err)
	assert.Equal(b, 5.0, resp.Rate, "Expected rate 5.0, got %f", resp.Rate)
	assert.Equal(b, "success", resp.Status, "Expected status 'success', got %s", resp.Status)
}
