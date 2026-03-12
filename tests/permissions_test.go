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

// TestGetPermissions_Empty verifies empty permissions response.
func TestGetPermissions_Empty(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	access, _ := exchangeGoogleTokens(
		t,
		ctx,
		st,
		googleMock,
		"perms_google_code",
		"testgoogle@example.com",
		"Google Test User",
	)

	req := connect.NewRequest(&authv1.GetPermissionsRequest{AccessToken: access})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	resp, err := st.AuthClient.GetPermissions(ctx, req)
	require.NoError(t, err)
	assert.Empty(t, resp.Msg.Permissions)
}

// TestGetPermissions_MissingAccessToken verifies failure when access token is missing.
func TestGetPermissions_MissingAccessToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.GetPermissionsRequest{})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.GetPermissions(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestGetPermissions_InvalidAccessToken verifies failure for invalid access token.
func TestGetPermissions_InvalidAccessToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.GetPermissionsRequest{AccessToken: "invalid-token"})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.GetPermissions(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
