// Package main contains the database connectivity and query logic for the record query microservice.
package main

import (
	"fmt"
	"log/slog"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// db holds the global database connection.
var db *sql.DB

// conectar establishes a connection to the MySQL database.
func conectar() error {
	// Define the MySQL connection string with credentials and database details.
	dataSource := fmt.Sprintf("dbuser:senha@tcp(%s:3306)/db?parseTime=true", "localhost")

	// Open a connection to the MySQL database.
	var err error
	db, err = sql.Open("mysql", dataSource)
	if err != nil {
		// Return a wrapped error if the connection fails.
		return fmt.Errorf("erro ao conectar ao MySQL: %w", err)
	}

	// Test the database connection with a ping.
	if err = db.Ping(); err != nil {
		// Return a wrapped error if the ping fails.
		return fmt.Errorf("erro ao pingar o MySQL: %w", err)
	}

	// Log a success message for the database connection.
	logger.Info("Conexão com MySQL estabelecida com sucesso")
	return nil
}

// consultar queries the database for records matching the provided ID.
func consultar(id string) (*Registros, error) {
	// Execute a SELECT query to retrieve records by ID.
	rows, err := db.Query("SELECT id, name FROM registros WHERE id = ?", id)
	if err != nil {
		// Return a wrapped error if the query fails.
		return nil, fmt.Errorf("erro ao consultar registros: %w", err)
	}
	// Ensure rows are closed after use.
	defer rows.Close()

	// Initialize a slice to store the query results.
	registros := make(Registros, 0)
	// Iterate over the query results.
	for rows.Next() {
		var r Registro
		// Scan the row into a Registro struct.
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			// Return a wrapped error if scanning fails.
			return nil, fmt.Errorf("erro ao escanear registro: %w", err)
		}
		// Append the record to the results slice.
		registros = append(registros, r)
	}

	// Check for any errors encountered during row iteration.
	if err := rows.Err(); err != nil {
		// Return a wrapped error if iteration fails.
		return nil, fmt.Errorf("erro ao iterar registros: %w", err)
	}

	// Log the number of records retrieved.
	logger.Info("Registros consultados com sucesso", slog.Int("count", len(registros)))
	// Return the results and no error.
	return &registros, nil
}
