package tests

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1"
	"github.com/salivare-io/sso-auth-server/tests/mocks"
	"github.com/salivare-io/sso-auth-server/tests/suite"
)

// TestRefreshToken_Rotation verifies refresh token rotation.
func TestRefreshToken_Rotation(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	_, refresh := exchangeGoogleTokens(
		t,
		ctx,
		st,
		googleMock,
		"refresh_google_code",
		"testgoogle@example.com",
		"Google Test User",
	)

	req := connect.NewRequest(&authv1.RefreshTokenRequest{RefreshToken: refresh})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	resp, err := st.AuthClient.RefreshToken(ctx, req)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Msg.AccessToken, "access token must not be empty")
	assert.NotEmpty(t, resp.Msg.RefreshToken, "refresh token must not be empty")
	assert.NotEqual(t, refresh, resp.Msg.RefreshToken, "refresh token must rotate")

	reqOld := connect.NewRequest(&authv1.RefreshTokenRequest{RefreshToken: refresh})
	reqOld.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)
	_, err = st.AuthClient.RefreshToken(ctx, reqOld)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

// TestRefreshToken_MissingToken verifies failure when refresh token is missing.
func TestRefreshToken_MissingToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.RefreshTokenRequest{})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.RefreshToken(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestRefreshToken_InvalidToken verifies failure for invalid refresh token.
func TestRefreshToken_InvalidToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.RefreshTokenRequest{RefreshToken: "invalid-refresh"})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.RefreshToken(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
