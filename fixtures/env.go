package fixtures

// This file contains the necessary imports and dependencies required for
// setting up and testing the environment for the AsyncAPI project. It includes
// packages for database interaction, configuration management, and testing utilities.

// Package imports:
// - "database/sql": Provides generic interface around SQL (or SQL-like) databases.
// - "fmt": Implements formatted I/O with functions analogous to C's printf and scanf.
// - "os": Provides functions to interact with the operating system, such as reading environment variables.
// - "strings": Contains functions to manipulate UTF-8 encoded strings.
// - "testing": Provides support for automated testing of Go packages.

// Project-specific imports:
// - "asyncapi/config": Handles configuration management for the AsyncAPI project.
// - "asyncapi/store": Manages data storage and retrieval for the AsyncAPI project.

// Third-party imports:
// - "github.com/golang-migrate/migrate/v4": A migration tool for managing database schema changes.
// - "github.com/lib/pq": A pure Go Postgres driver for database/sql.
// - "github.com/stretchr/testify/require": Provides assertion methods for testing, ensuring test conditions are met.
import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"asyncapi/config"
	"asyncapi/store"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestEnv represents the test environment for integration tests.
// It provides the configuration and database connection required for testing.
type TestEnv struct {
	Config *config.Config // The application configuration for the test environment.
	Db     *sql.DB        // The database connection used for testing.
}

// NewTestEnv initializes a new test environment for integration tests.
//
// This function sets the environment to "test", creates a new configuration,
// and establishes a connection to the PostgreSQL database.
//
// Parameters:
// - t: The testing object used for assertions and cleanup.
//
// Returns:
// - A pointer to a TestEnv instance containing the configuration and database connection.
func NewTestEnv(t *testing.T) *TestEnv {
	// Set the environment variable to indicate the test environment
	os.Setenv("ENV", string(config.Env_Test))

	// Create a new configuration for the test environment
	conf, err := config.New()
	if err != nil {
		fmt.Println("Error creating config:", err)
		return nil
	}

	// Print debug information about the database URL and project root
	fmt.Printf(">>> database url: %s\n", conf.DatabaseUrl())
	fmt.Printf(">>> project root: %s\n", os.Getenv("PROJECT_ROOT"))

	// Establish a connection to the PostgreSQL database
	db, err := store.NewPostgresDb(conf)
	require.NoError(t, err)
	if err != nil {
		fmt.Println("Error creating DB:", err)
		return nil
	}

	// Return the initialized test environment
	return &TestEnv{
		Db:     db,
		Config: conf,
	}
}

// SetupDb applies database migrations and prepares the database for testing.
//
// This function runs all the migrations in the `migrations` directory to ensure
// the database schema is up-to-date. It also returns a teardown function to clean up
// the database after the tests are completed.
//
// Parameters:
// - t: The testing object used for assertions and cleanup.
//
// Returns:
// - A function that can be called to clean up the database after tests.
func (te *TestEnv) SetupDb(t *testing.T) func(t *testing.T) {
	// Initialize the migration tool with the migrations directory and database URL
	m, err := migrate.New(
		fmt.Sprintf("file:///%s/migrations", te.Config.ProjectRoot),
		te.Config.DatabaseUrl(),
	)
	require.NoError(t, err)

	// Apply all migrations to set up the database schema
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	// Return the teardown function to clean up the database after tests
	return te.TeardownDb
}

// TeardownDb cleans up the database after tests.
//
// This function truncates all the tables in the database, closes the database connection,
// and ensures the test database is properly cleaned up.
//
// Parameters:
// - t: The testing object used for assertions and cleanup.
func (te *TestEnv) TeardownDb(t *testing.T) {
	// Truncate all tables to remove test data
	_, err := te.Db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join([]string{"users", "refresh_tokens", "reports"}, ",")))
	require.NoError(t, err)

	// Close the database connection
	if err := te.Db.Close(); err != nil {
		require.NoError(t, err)
	}

	// Drop the test database (commented out for safety)
	// Uncomment the following lines if you want to drop the test database after tests
	// _, err = te.Db.Exec("DROP DATABASE %s", te.Config.DatabaseName)
	// require.NoError(t, err)
}
