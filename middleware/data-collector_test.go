// Package middleware provides tests for the middleware functions of the pricing microservice.
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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestDataEnrichment tests the DataEnrichment middleware for various scenarios.
func TestDataEnrichment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		// Mock external API
		var receivedCorrelationID string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedCorrelationID = r.Header.Get("x-app-correlationID")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"credit_score": 750.0}`))
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

		var capturedContext *gin.Context
		router := gin.New()
		router.Use(func(c *gin.Context) {
			capturedContext = c // Capture context before DataEnrichment
			c.Next()
		})
		router.Use(DataEnrichment())
		router.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		reqBody, _ := json.Marshal(model.PricingRequest{
			LoanAmount: 10000.0,
			CustomerID: "CUST12345",
		})
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		t.Logf("Request body: %s", string(reqBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
		assert.NotNil(t, capturedContext, "Expected captured context")
		if w.Code == http.StatusOK {
			enrichedData, exists := capturedContext.Get("enriched_data")
			assert.True(t, exists, "Expected enriched_data in context")
			assert.NotNil(t, enrichedData, "Expected non-nil enriched_data")
			data, ok := enrichedData.(map[string]interface{})
			assert.True(t, ok, "Expected enriched_data to be map[string]interface{}")
			assert.Equal(t, 750.0, data["credit_score"], "Expected credit_score 750, got %v", data["credit_score"])
			correlationID, exists := capturedContext.Get("correlation_id")
			assert.True(t, exists, "Expected correlation_id in context")
			assert.Equal(t, "test-correlation-id-123", correlationID, "Expected correlation_id test-correlation-id-123, got %v", correlationID)
			assert.Equal(t, "test-correlation-id-123", receivedCorrelationID, "Expected x-app-correlationID header test-correlation-id-123, got %s", receivedCorrelationID)
		}
	})

	t.Run("MissingCorrelationID", func(t *testing.T) {
		// Mock external API
		var receivedCorrelationID string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedCorrelationID = r.Header.Get("x-app-correlationID")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"credit_score": 750.0}`))
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

		var capturedContext *gin.Context
		router := gin.New()
		router.Use(func(c *gin.Context) {
			capturedContext = c // Capture context before DataEnrichment
			c.Next()
		})
		router.Use(DataEnrichment())
		router.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		reqBody, _ := json.Marshal(model.PricingRequest{
			LoanAmount: 10000.0,
			CustomerID: "CUST12345",
		})
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// No x-app-correlationID header
		w := httptest.NewRecorder()

		t.Logf("Request body: %s", string(reqBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
		assert.NotNil(t, capturedContext, "Expected captured context")
		if w.Code == http.StatusOK {
			correlationID, exists := capturedContext.Get("correlation_id")
			assert.True(t, exists, "Expected correlation_id in context")
			assert.Equal(t, "unknown", correlationID, "Expected correlation_id unknown, got %v", correlationID)
			assert.Equal(t, "unknown", receivedCorrelationID, "Expected x-app-correlationID header unknown, got %s", receivedCorrelationID)
		}
	})

	t.Run("InvalidCustomerID", func(t *testing.T) {
		var capturedContext *gin.Context
		router := gin.New()
		router.Use(func(c *gin.Context) {
			capturedContext = c // Capture context before DataEnrichment
			c.Next()
		})
		router.Use(DataEnrichment())
		router.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		reqBody, _ := json.Marshal(model.PricingRequest{
			LoanAmount: 10000.0,
			CustomerID: "",
		})
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		t.Logf("Request body: %s", string(reqBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status 500, got %d: %s", w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "failed to enrich data: invalid customer ID or loan amount")
		assert.NotNil(t, capturedContext, "Expected captured context")
		// Check correlation_id even on abort
		correlationID, exists := capturedContext.Get("correlation_id")
		assert.True(t, exists, "Expected correlation_id in context")
		assert.Equal(t, "test-correlation-id-123", correlationID, "Expected correlation_id test-correlation-id-123, got %v", correlationID)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		var capturedContext *gin.Context
		router := gin.New()
		router.Use(func(c *gin.Context) {
			capturedContext = c // Capture context before DataEnrichment
			c.Next()
		})
		router.Use(DataEnrichment())
		router.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		t.Logf("Request body: %s", "invalid json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400, got %d: %s", w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "invalid request")
		assert.NotNil(t, capturedContext, "Expected captured context")
		// Check correlation_id even on abort
		correlationID, exists := capturedContext.Get("correlation_id")
		assert.True(t, exists, "Expected correlation_id in context")
		assert.Equal(t, "test-correlation-id-123", correlationID, "Expected correlation_id test-correlation-id-123, got %v", correlationID)
	})

	t.Run("CorrelationIDPropagation", func(t *testing.T) {
		// Mock external API
		var receivedCorrelationID string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedCorrelationID = r.Header.Get("x-app-correlationID")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"credit_score": 750.0}`))
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

		var capturedContext *gin.Context
		router := gin.New()
		router.Use(func(c *gin.Context) {
			capturedContext = c // Capture context before DataEnrichment
			c.Next()
		})
		router.Use(DataEnrichment())
		router.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		reqBody, _ := json.Marshal(model.PricingRequest{
			LoanAmount: 10000.0,
			CustomerID: "CUST12345",
		})
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-app-correlationID", "test-correlation-id-123")
		w := httptest.NewRecorder()

		t.Logf("Request body: %s", string(reqBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
		assert.Equal(t, "test-correlation-id-123", receivedCorrelationID, "Expected x-app-correlationID header test-correlation-id-123, got %s", receivedCorrelationID)
		assert.NotNil(t, capturedContext, "Expected captured context")
		if w.Code == http.StatusOK {
			correlationID, exists := capturedContext.Get("correlation_id")
			assert.True(t, exists, "Expected correlation_id in context")
			assert.Equal(t, "test-correlation-id-123", correlationID, "Expected correlation_id test-correlation-id-123, got %v", correlationID)
		}
	})
}
