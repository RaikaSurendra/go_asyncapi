package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// The version of the AsyncAPI specification that the document uses.
	// REQUIRED.
	// Example: 2.0.0
	DatabaseName            string `env:"DB_NAME" envDefault:"asyncapi"`
	DatabaseHost            string `env:"DB_HOST" envDefault:"localhost"`
	DatabasePort            int    `env:"DB_PORT" envDefault:"5432"`
	DatabaseUser            string `env:"DB_USER" envDefault:"postgres"`
	DatabasePassword        string `env:"DB_PASSWORD" envDefault:"password"`
	DatabaseSSLMode         string `env:"DB_SSL_MODE" envDefault:"disable"`
	DatabaseMaxIdleConns    int    `env:"DB_MAX_IDLE_CONNS" envDefault:"10"`
	DatabaseMaxOpenConns    int    `env:"DB_MAX_OPEN_CONNS" envDefault:"100"`
	DatabaseConnMaxLifetime int    `env:"DB_CONN_MAX_LIFETIME" envDefault:"300"`
	DatabaseConnMaxIdleTime int    `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"300"`
	DatabaseURL             string `env:"DATABASE_URL"`
}

func NewConfig() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
