// Package model defines data structures for the pricing microservice.
package model

// PricingRequest represents the input data for a loan pricing request.
type PricingRequest struct {
	LoanAmount float64 `json:"loan_amount"`
	CustomerID string  `json:"customer_id"`
}

// PricingResponse represents the response data for a pricing request.
type PricingResponse struct {
	Rate   float64 `json:"rate"`
	Status string  `json:"status"`
	Error  string  `json:"error,omitempty"`
}
