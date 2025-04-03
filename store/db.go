package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"asyncapi/config"

	_ "github.com/lib/pq" // PostgreSQL driver
	// Import the PostgreSQL driver
	// to register it with the database/sql package
	// This is necessary to use the PostgreSQL driver with the sql package
	// without directly referencing it in the code.
	// The underscore import is a common Go idiom
	// to indicate that the package is imported for its side effects
	// (in this case, registering the driver)
	// and not for any exported identifiers.
)

func NewPostgresDb(conf *config.Config) (*sql.DB, error) {
	// Create a connection string using the configuration
	dsn := conf.DatabaseUrl()
	// Open a new database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	//Context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Set the maximum number of open connections to the database
	db.SetMaxOpenConns(10)
	// Set the maximum number of idle connections to the database
	db.SetMaxIdleConns(5)
	// Set the maximum lifetime of a connection to the database
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping the database to ensure it's reachable
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
