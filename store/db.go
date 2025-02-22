package store

import (
	"database/sql"
	"fmt"
	"log"
	"asyncapi/config"
	_ "github.com/lib/pq"
)

func NewPostgresStore(config *config.Config) *PostgresStore {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatal(, fmt.Errorf("failed to open database: %w", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(fmt.Errorf("failed to ping database: %w", err))
	}
	return &PostgresStore{db: db}
}