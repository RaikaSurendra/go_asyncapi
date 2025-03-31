package store

import (
	"context"
	"fmt"
	"os"
	"testing"

	"asyncapi/config"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	os.Setenv("ENV", string(config.Env_Test))
	// Replace with your test database connection string
	conf, err := config.New()
	require.NoError(t, err)
	require.NotNil(t, conf)
	fmt.Printf(">>> database url: %s\n", conf.DatabaseUrl())

	db, err := NewPostgresDb(conf)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Return the database connection if needed for other tests
	_ = db
	// Ensure the database is cleaned up after tests
	m, err := migrate.New(
		fmt.Sprintf("file:///%s/migrations", conf.ProjectRoot),
		conf.DatabaseUrl(),
	)
	require.NoError(t, err)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	userStore := NewUserStore(db)
	require.NotNil(t, userStore)
	// Create a test user
	ctx := context.Background()
	user, err := userStore.CreateUser(ctx, "test@test.com", "testpassword")
	require.NoError(t, err)
	require.NotNil(t, user)

	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("testpassword"))

	// Clean up the database after tests

}
