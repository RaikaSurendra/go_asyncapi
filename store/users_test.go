package store_test

import (
	"context"
	"testing"
	"time"

	"asyncapi/fixtures"
	"asyncapi/store"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {

	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	require.NotNil(t, cleanup)
	// Ensure the database is cleaned up after tests
	t.Cleanup(func() {
		cleanup(t)
	})
	userStore := store.NewUserStore(env.Db)
	require.NotNil(t, userStore)
	now := time.Now()
	// Create a test user
	ctx := context.Background()
	user, err := userStore.CreateUser(ctx, "test@test.com", "testpassword")
	require.NoError(t, err)
	require.NotNil(t, user)

	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("testpassword"))
	require.Less(t, now.UnixNano(), user.CreatedAt.UnixNano())

	// Clean up the database after tests
	user2, err := userStore.GetUserByID(ctx, user.Id)
	require.NoError(t, err)
	require.NotNil(t, user2.Email)
	require.Equal(t, user.Email, user2.Email)
	require.Equal(t, user.Id, user2.Id)
	require.Equal(t, user.HashedPasswordBase64, user2.HashedPasswordBase64)
	require.Equal(t, user.CreatedAt.UnixNano(), user2.CreatedAt.UnixNano())

	// Clean up the database after tests
	user2, err = userStore.GetUserByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.NotNil(t, user2.Email)
	require.Equal(t, user.Email, user2.Email)
	require.Equal(t, user.Id, user2.Id)
	require.Equal(t, user.HashedPasswordBase64, user2.HashedPasswordBase64)
	require.Equal(t, user.CreatedAt.UnixNano(), user2.CreatedAt.UnixNano())

}
