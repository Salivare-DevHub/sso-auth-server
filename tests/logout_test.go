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

// TestLogout_InvalidatesRefreshToken verifies logout invalidates refresh token.
func TestLogout_InvalidatesRefreshToken(t *testing.T) {
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
		"logout_google_code",
		"testgoogle@example.com",
		"Google Test User",
	)

	req := connect.NewRequest(&authv1.LogoutRequest{RefreshToken: refresh})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.Logout(ctx, req)
	require.NoError(t, err)

	reqOld := connect.NewRequest(&authv1.RefreshTokenRequest{RefreshToken: refresh})
	reqOld.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)
	_, err = st.AuthClient.RefreshToken(ctx, reqOld)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

// TestLogout_MissingToken verifies failure when refresh token is missing.
func TestLogout_MissingToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.LogoutRequest{})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.Logout(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestLogout_InvalidToken verifies failure for invalid refresh token.
func TestLogout_InvalidToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.LogoutRequest{RefreshToken: "invalid-refresh"})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.Logout(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
