package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseName     string `env:"DB_NAME"`
	DatabaseUser     string `env:"DB_USER"`
	DatabasePassword string `env:"DB_PASSWORD"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabaseSSLMode  string `env:"DB_SSL_MODE"`
}

func (c *Config) DatabaseUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName, c.DatabaseSSLMode)
}

func New() (*Config, error) {

	os.WriteFile(".env", []byte(`
		    DB_NAME=asyncapi
		    DB_USER=admin
		    DB_PASSWORD=secret
		    DB_HOST=127.0.0.1
		    DB_PORT=5432
		    DB_SSL_MODE=disable
		    `), 0644)

	// set os environment variables from .envrc
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Parse environment variables into Config struct
	// Use env.Parse to parse the environment variables into the Config struct
	// The env package will automatically look for the environment variables
	// defined in the struct tags (e.g., DB_NAME, DB_USER, etc.)
	// If any of the required environment variables are missing or invalid,
	// env.Parse will return an error.
	// The env package will automatically look for the environment variables
	// defined in the struct tags (e.g., DB_NAME, DB_USER, etc.)
	// If any of the required environment variables are missing or invalid,
	// env.Parse will return an error.
	// The env package will automatically look for the environment variables
	// defined in the struct tags (e.g., DB_NAME, DB_USER, etc.)

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return &cfg, nil
}
