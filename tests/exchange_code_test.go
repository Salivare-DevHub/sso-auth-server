package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1"
	"github.com/salivare-io/sso-auth-server/tests/mocks"
	"github.com/salivare-io/sso-auth-server/tests/suite"
)

// setupTestOAuthEnv sets env vars for OAuth tests.
func setupTestOAuthEnv(t *testing.T, googleMockURL, yandexMockURL string) {
	t.Helper()

	oauthMap := map[string]map[string]interface{}{
		"google": {
			"client_id":     "test-google-client-id",
			"client_secret": "test-google-secret",
			"redirect_url":  "http://localhost/callback",
			"token_url":     fmt.Sprintf("%s/token", googleMockURL),
			"userinfo_url":  fmt.Sprintf("%s/v1/userinfo", googleMockURL),
			"scopes": []string{
				"openid",
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
		},
		"yandex": {
			"client_id":     "test-yandex-client-id",
			"client_secret": "test-yandex-secret",
			"redirect_url":  "http://localhost/callback",
			"token_url":     fmt.Sprintf("%s/token", yandexMockURL),
			"userinfo_url":  fmt.Sprintf("%s/info", yandexMockURL),
			"scopes": []string{
				"login:email",
				"login:info",
			},
		},
	}
	internalMap := map[string]map[string]interface{}{
		"auth_client": {
			"client_id":    "auth-client-id",
			"bearer_token": "billing-secret",
			"signing_key":  "test-signing-key",
			"access_ttl":   "15m",
			"refresh_ttl":  "168h", // 7 days
		},
	}

	oauthJSON, err := json.Marshal(oauthMap)
	require.NoError(t, err)
	internalJSON, err := json.Marshal(internalMap)
	require.NoError(t, err)

	require.NoError(t, os.Setenv("OAUTH_PROVIDERS", string(oauthJSON)))
	require.NoError(t, os.Setenv("INTERNAL_APPS", string(internalJSON)))
}

// cleanupTestOAuthEnv clears env vars after a test.
func cleanupTestOAuthEnv(t *testing.T) {
	t.Helper()
	_ = os.Unsetenv("OAUTH_PROVIDERS")
	_ = os.Unsetenv("INTERNAL_APPS")
}

// newExchangeCodeRequest builds a request to exchange an auth code.
func newExchangeCodeRequest(
	t *testing.T,
	provider authv1.Provider,
	code string,
) *connect.Request[authv1.ExchangeCodeRequest] {
	t.Helper()

	req := connect.NewRequest(
		&authv1.ExchangeCodeRequest{
			Provider: provider,
			Code:     code,
		},
	)

	return req
}

// TestExchangeCode_Google verifies successful Google code exchange.
func TestExchangeCode_Google(t *testing.T) {
	// Start the Google mock server.
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	// Set env vars with mock server URLs.
	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	// Initialize the suite (loads config from env vars).
	ctx, st := suite.New(t)

	// Register a test code in Google mock.
	testCode := "test_google_code_123"
	googleMock.RegisterCode(
		testCode,
		mocks.GoogleUserInfo{
			ID:            "google_user_123",
			Email:         "testgoogle@example.com",
			VerifiedEmail: true,
			Name:          "Google Test User",
			Picture:       "https://example.com/pic.jpg",
		},
	)

	// Create request with bearer token from config.
	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_GOOGLE, testCode)
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	// Execute request.
	resp, err := st.AuthClient.ExchangeCode(ctx, req)
	require.NoError(t, err)

	// Verify response.
	assert.NotEmpty(t, resp.Msg.AccessToken, "access token must not be empty")
	assert.NotEmpty(t, resp.Msg.RefreshToken, "refresh token must not be empty")
}

// TestExchangeCode_Yandex verifies successful Yandex code exchange.
func TestExchangeCode_Yandex(t *testing.T) {
	// Start mock servers.
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	// Set env vars with mock server URLs.
	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	// Initialize the suite.
	ctx, st := suite.New(t)

	// Register a test code in Yandex mock.
	testCode := "test_yandex_code_456"
	yandexMock.RegisterCode(
		testCode,
		mocks.YandexUserInfo{
			ID:        987654321,
			Login:     "testyandex",
			FirstName: "Yandex",
			LastName:  "User",
			RealName:  "Yandex Test User",
			Email:     "testyandex@example.com",
			Picture:   "avatar_id_123",
		},
	)

	// Create request with bearer token from config.
	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_YANDEX, testCode)
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	// Execute request.
	resp, err := st.AuthClient.ExchangeCode(ctx, req)
	require.NoError(t, err)

	// Verify response.
	assert.NotEmpty(t, resp.Msg.AccessToken, "access token must not be empty")
	assert.NotEmpty(t, resp.Msg.RefreshToken, "refresh token must not be empty")
}

// TestExchangeCode_MissingAuthHeader verifies failure when Authorization is missing.
func TestExchangeCode_MissingAuthHeader(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_GOOGLE, "missing_auth_code")

	_, err := st.AuthClient.ExchangeCode(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

// TestExchangeCode_InvalidBearerToken verifies failure for invalid bearer token.
func TestExchangeCode_InvalidBearerToken(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_GOOGLE, "invalid_token_code")
	req.Header().Set("Authorization", "Bearer not-a-real-token")

	_, err := st.AuthClient.ExchangeCode(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

// TestExchangeCode_InvalidProvider verifies failure for an unsupported provider.
func TestExchangeCode_InvalidProvider(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_UNSPECIFIED, "any_code")
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.ExchangeCode(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// TestExchangeCode_InvalidCode verifies failure for an invalid OAuth code.
func TestExchangeCode_InvalidCode(t *testing.T) {
	googleMock := mocks.NewGoogleServerMock()
	defer googleMock.Close()

	yandexMock := mocks.NewYandexServerMock()
	defer yandexMock.Close()

	setupTestOAuthEnv(t, googleMock.URL(), yandexMock.URL())
	defer cleanupTestOAuthEnv(t)

	ctx, st := suite.New(t)

	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_GOOGLE, "unknown_code")
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	_, err := st.AuthClient.ExchangeCode(ctx, req)
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}
