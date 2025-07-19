// Package main defines the data structures for the record query microservice.
package main

// Registro represents the structure of a record in the MySQL registros table.
type Registro struct {
	// ID is the unique identifier for the record.
	ID string `json:"id"`
	// Name is the name associated with the record.
	Name string `json:"name"`
}

// Registros is a slice of Registro, used to store multiple records.
type Registros []Registro

// Payload represents the structure of the input data for API queries.
type Payload struct {
	// UserID is the required identifier for the user making the query.
	UserID string `json:"user_id" binding:"required"`
	// Name is an optional field for additional query data.
	Name string `json:"name,omitempty"`
}
