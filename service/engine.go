// Package service provides business logic for the pricing microservice.
// This file implements the integration with an external decision engine.
package service

import (
	"context"
	"net/http"
	"time"

	"github.com/raywall/go-api-middleware-service/model"
	"github.com/sony/gobreaker"
)

// DecisionEngine handles communication with an external decision engine.
type DecisionEngine struct {
	client         *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

// NewDecisionEngine creates a new DecisionEngine with a configured HTTP client and circuit breaker.
func NewDecisionEngine() *DecisionEngine {
	return &DecisionEngine{
		client: &http.Client{Timeout: 5 * time.Second},
		circuitBreaker: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "decision_engine",
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     10 * time.Second,
		}),
	}
}

// CalculateRate calls the external decision engine to compute the pricing rate.
// It uses a circuit breaker to ensure resilience against failures.
func (de *DecisionEngine) CalculateRate(ctx context.Context, req model.PricingRequest, enrichedData map[string]interface{}) (float64, error) {
	result, err := de.circuitBreaker.Execute(func() (interface{}, error) {
		// Chamada HTTP ao motor de decisão
		// Exemplo: POST /decision-engine/rate
		rate := 5.0 // Mock
		return rate, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
