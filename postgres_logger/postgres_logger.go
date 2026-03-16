package postgres_logger

import (
	"Monitor/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

type PostgresLogger struct {
	ctx        context.Context
	wg         *sync.WaitGroup
	cfg        *config.Config
	mutex      *sync.RWMutex
	connString string
}

func NewPostgresLogger(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, mutex *sync.RWMutex) *PostgresLogger {
	mutex.RLock()
	user := cfg.PostgresConfig.User
	password := cfg.PostgresConfig.Password
	host := cfg.PostgresConfig.Host
	port := cfg.PostgresConfig.Port
	database := cfg.PostgresConfig.Database
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)
	mutex.RUnlock()

	return &PostgresLogger{
		ctx:        ctx,
		wg:         wg,
		cfg:        cfg,
		mutex:      mutex,
		connString: connString,
	}
}

func (pl *PostgresLogger) Info(value string) error {

	tableName := "strings"

	db, err := sql.Open("postgres", pl.connString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s (string) VALUES ($1)", tableName)
	_, err = db.Exec(query, value)
	if err != nil {
		return fmt.Errorf("failed to insert string: %w", err)
	}

	return nil
}
