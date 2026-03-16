package postgres_logger

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresLogger struct{}

func NewPostgresLogger() *PostgresLogger {
	return &PostgresLogger{}
}

// StoreString stores a string value in a PostgreSQL database
// connectionString format: "host=localhost port=5432 user=username password=password dbname=mydb sslmode=disable"
// tableName: the name of the table to store the string (must exist with a 'value' TEXT column)
func StoreString(connectionString, tableName, value string) error {
	// Open database connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Insert the string into the database
	query := fmt.Sprintf("INSERT INTO %s (string) VALUES ($1)", tableName)
	_, err = db.Exec(query, value)
	if err != nil {
		return fmt.Errorf("failed to insert string: %w", err)
	}

	return nil
}
