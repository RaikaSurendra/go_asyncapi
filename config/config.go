package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Env string

const (
	Env_Test Env = "test"
	Env_Dev  Env = "dev"
	Env_Prod Env = "prod"
)

type Config struct {
	ApiServerPort        string `env:"APISERVER_PORT"`
	ApiServerHost        string `env:"APISERVER_HOST"`
	DatabaseName         string `env:"DB_NAME"`
	DatabaseUser         string `env:"DB_USER"`
	DatabasePassword     string `env:"DB_PASSWORD"`
	DatabaseHost         string `env:"DB_HOST"`
	DatabasePort         string `env:"DB_PORT"`
	DatabasePortTest     string `env:"DB_PORT_TEST"`
	DatabaseSSLMode      string `env:"DB_SSL_MODE"`
	Env                  Env    `env:"ENV" envDefault:"dev"`
	ProjectRoot          string `env:"PROJECT_ROOT" envDefault:"/Users/surendraraika/projects/asyncapi"`
	JwtSecret            string `env:"JWT_SECRET"`
	S3LocalstackEndpoint string `env:"S3_LOCALSTACK_ENDPOINT"`
	ReportsSQSEndpoint   string `env:"REPORTS_SQS_ENDPOINT"`
	S3Bucket             string `env:"S3_BUCKET"`
	SqsQueue             string `env:"SQS_QUEUE"`
}

func (c *Config) DatabaseUrl() string {
	port := c.DatabasePort
	if c.Env == Env_Test {
		port = c.DatabasePortTest
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, port, c.DatabaseName, c.DatabaseSSLMode)
}

func New() (*Config, error) {

	wd := "/Users/surendraraika/projects/asyncapi"
	if wd == "" {
		return nil, fmt.Errorf("failed to get working directory: PROJECT_HOME is not set")
	}

	os.WriteFile(".env", fmt.Appendf(nil, `
			JWT_SECRET=supersecret
			APISERVER_PORT=5001
			APISERVER_HOST=localhost
			DB_NAME=asyncapi
			DB_USER=admin
			DB_PASSWORD=secret
			DB_HOST=127.0.0.1
			DB_PORT=5432
			DB_PORT_TEST=5433
			DB_SSL_MODE=disable
			PROJECT_ROOT=%s
			`, wd), 0644)

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
