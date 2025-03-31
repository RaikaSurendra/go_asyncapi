package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {

	// Call the New function to parse the environment variables
	cfg, err := New()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	fmt.Println("Config:", cfg)

	// Assert that the configuration is correctly populated

	require.Equal(t, "asyncapi", cfg.DatabaseName)
	require.Equal(t, "admin", cfg.DatabaseUser)
	require.Equal(t, "secret", cfg.DatabasePassword)
	require.Equal(t, "127.0.0.1", cfg.DatabaseHost)
	require.Equal(t, "5432", cfg.DatabasePort)
	require.Equal(t, "disable", cfg.DatabaseSSLMode)
	require.Equal(t, "5433", cfg.DatabasePortTest)
	require.Equal(t, Env_Dev, cfg.Env)

}
func TestDatabaseUrl(t *testing.T) {
	cfg := &Config{
		DatabaseName:     "asyncapi",
		DatabaseUser:     "admin",
		DatabasePassword: "secret",
		DatabaseHost:     "127.0.0.1",
		DatabasePort:     "5432",
		DatabasePortTest: "5433",
		DatabaseSSLMode:  "disable",
		ProjectRoot:      "/Users/surendraraika/projects/asyncapi",
	}

	// Test for development environment
	cfg.Env = Env_Dev
	expectedDevUrl := "postgres://admin:secret@127.0.0.1:5432/asyncapi?sslmode=disable"
	require.Equal(t, expectedDevUrl, cfg.DatabaseUrl())

	// Test for test environment
	cfg.Env = Env_Test
	expectedTestUrl := "postgres://admin:secret@127.0.0.1:5433/asyncapi?sslmode=disable"
	require.Equal(t, expectedTestUrl, cfg.DatabaseUrl())

	// Test for production environment
	cfg.Env = Env_Prod
	expectedProdUrl := "postgres://admin:secret@127.0.0.1:5432/asyncapi?sslmode=disable"
	require.Equal(t, expectedProdUrl, cfg.DatabaseUrl())
}
