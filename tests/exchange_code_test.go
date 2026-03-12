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
