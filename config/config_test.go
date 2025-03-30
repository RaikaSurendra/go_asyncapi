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
	require.Equal(t, 5432, cfg.DatabasePort)
	require.Equal(t, "disable", cfg.DatabaseSSLMode)

}
