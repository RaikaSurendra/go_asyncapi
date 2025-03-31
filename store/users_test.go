package store

import (
	"context"
	"testing"

	"github.com/RaikaSurendra/go_asyncapi/fixtures"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})
	userStore := NewUserStore(env.Db)
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
