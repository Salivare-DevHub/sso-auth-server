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

// TestRevokeRefreshToken_InvalidatesRefreshToken verifies revoke invalidates refresh token.
func TestRevokeRefreshToken_InvalidatesRefreshToken(t *testing.T) {
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
		"revoke_google_code",
		"testgoogle@example.com",
		"Google Test User",
	)

	req := connect.NewRequest(&authv1.RevokeRefreshTokenRequest{RefreshToken: refresh})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.RevokeRefreshToken(ctx, req)
	require.NoError(t, err)

	reqOld := connect.NewRequest(&authv1.RefreshTokenRequest{RefreshToken: refresh})
	reqOld.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)
	_, err = st.AuthClient.RefreshToken(ctx, reqOld)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

// TestRevokeRefreshToken_MissingToken verifies failure when refresh token is missing.
func TestRevokeRefreshToken_MissingToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.RevokeRefreshTokenRequest{})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.RevokeRefreshToken(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestRevokeRefreshToken_InvalidToken verifies failure for invalid refresh token.
func TestRevokeRefreshToken_InvalidToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.RevokeRefreshTokenRequest{RefreshToken: "invalid-refresh"})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.RevokeRefreshToken(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
