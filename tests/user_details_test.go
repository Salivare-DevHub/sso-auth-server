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

// TestUserDetails_Google verifies user details response for valid access token.
func TestUserDetails_Google(t *testing.T) {
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
		"details_google_code",
		"testgoogle@example.com",
		"Google Test User",
	)

	req := connect.NewRequest(&authv1.UserDetailsRequest{AccessToken: access})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	resp, err := st.AuthClient.UserDetails(ctx, req)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Msg.UserId, "user_id must not be empty")
	assert.Equal(t, "testgoogle@example.com", resp.Msg.Email)
	assert.Equal(t, "Google Test User", resp.Msg.Name)
}

// TestUserDetails_MissingAccessToken verifies failure when access token is missing.
func TestUserDetails_MissingAccessToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.UserDetailsRequest{})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.UserDetails(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestUserDetails_InvalidAccessToken verifies failure for invalid access token.
func TestUserDetails_InvalidAccessToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := connect.NewRequest(&authv1.UserDetailsRequest{AccessToken: "invalid-token"})
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.UserDetails(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
