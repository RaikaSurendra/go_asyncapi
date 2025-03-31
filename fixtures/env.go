package fixtures

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"strings"

	"github.com/RaikaSurendra/go_asyncapi/config"
	"github.com/RaikaSurendra/go_asyncapi/store"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	Config *config.Config
	Db     *sql.DB
}

func NewTestEnv(t *testing.T) *TestEnv {
	// Set the environment variable for testing
	os.Setenv("ENV", string(config.Env_Test))

	// Create a new configuration
	conf, err := config.New()
	if err != nil {
		fmt.Println("Error creating config:", err)
		return nil
	}

	fmt.Printf(">>> database url: %s\n", conf.DatabaseUrl())
	fmt.Printf(">>> project root: %s\n", os.Getenv("PROJECT_ROOT"))

	// Create a new PostgreSQL database connection
	db, err := store.NewPostgresDb(conf)
	require.NoError(t, err)
	if err != nil {
		fmt.Println("Error creating DB:", err)
		return nil
	}

	return &TestEnv{
		Db:     db,
		Config: conf}
}

func (te *TestEnv) SetupDb(t *testing.T) {
	// Ensure the database is cleaned up after tests
	m, err := migrate.New(
		fmt.Sprintf("file:///%s/migrations", os.Getenv("PROJECT_ROOT")),
		te.Config.DatabaseUrl(),
	)
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

func (te *TestEnv) TeardownDb(t *testing.T) {
	_, err := te.Db.Exec("TRUNCATE TABLE %s CASCADE", strings.Join([]string{"users", "refresh_tokens", "reports"}, ","))
	require.NoError(t, err)
	// Close the database connection
	if err := te.Db.Close(); err != nil {
		require.NoError(t, err)
	}
	// Drop the test database
	// _, err = te.Db.Exec("DROP DATABASE %s", te.Config.DatabaseName)
	// require.NoError(t, err)
}
