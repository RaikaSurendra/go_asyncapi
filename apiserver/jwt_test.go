package apiserver_test

import (
	"testing"

	"asyncapi/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"asyncapi/apiserver" // Import the correct package for apiServer
)

// TestJwtManager_GenerateTokenPair verifies the correctness of the generated token pair
//   - Verifies that the subject of the access token is the same as the given userId
//   - Verifies that the issuer of the access token is the same as the ApiServerHost and ApiServerPort
//   - Verifies that the subject of the refresh token is the same as the given userId
//   - Verifies that the issuer of the refresh token is the same as the ApiServerHost and ApiServerPort
//   - Verifies that the parsed token is the same as the original token
func TestJwtManager_GenerateTokenPair(t *testing.T) {
	// Mock configuration
	mockConfig, err := config.New()
	require.NoError(t, err)

	jwtManager := apiserver.NewJwtManager(mockConfig)
	userId := uuid.New()
	tokenPair, err := jwtManager.GenerateTokenPair(userId)
	require.NoError(t, err)

	//test isAccessToken method
	require.True(t, jwtManager.IsAccessToken(tokenPair.AccessToken))
	require.False(t, jwtManager.IsAccessToken(tokenPair.RefreshToken))

	subject, err := tokenPair.AccessToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), subject)

	accessTokenIssuer, err := tokenPair.AccessToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, "http://"+mockConfig.ApiServerHost+":"+mockConfig.ApiServerPort, accessTokenIssuer)

	refreshTokenSubject, err := tokenPair.RefreshToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), refreshTokenSubject)

	refreshTokenIssuer, err := tokenPair.RefreshToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, "http://"+mockConfig.ApiServerHost+":"+mockConfig.ApiServerPort, refreshTokenIssuer)

	parsedAccessToken, err := jwtManager.Parse(tokenPair.AccessToken.Raw)
	require.NoError(t, err)
	require.Equal(t, tokenPair.AccessToken, parsedAccessToken)

	parsedrefreshToken, err := jwtManager.Parse(tokenPair.RefreshToken.Raw)
	require.NoError(t, err)
	require.Equal(t, tokenPair.RefreshToken, parsedrefreshToken)

}
