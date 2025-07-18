// Package service provides tests for the business logic of the pricing microservice.
package service

import (
	"context"
	"testing"

	"github.com/raywall/go-api-middleware-service/model"

	"github.com/stretchr/testify/assert"
)

// TestCalculateRate tests the CalculateRate method of DecisionEngine.
func TestCalculateRate(t *testing.T) {
	engine := NewDecisionEngine()

	t.Run("Success", func(t *testing.T) {
		req := model.PricingRequest{
			LoanAmount: 10000.0,
			CustomerID: "CUST12345",
		}
		data := map[string]interface{}{"credit_score": 750}

		rate, err := engine.CalculateRate(context.Background(), req, data)

		assert.NoError(t, err)
		assert.Equal(t, 5.0, rate)
	})

	// Note: Testing circuit breaker failure requires mocking gobreaker.Execute
	// This is omitted as the current implementation uses a mock response
}
