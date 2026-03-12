package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"connectrpc.com/connect"
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

func exchangeGoogleTokens(
	t *testing.T,
	ctx context.Context,
	st *suite.Suite,
	googleMock *mocks.GoogleServerMock,
	code string,
	email string,
	name string,
) (string, string) {
	t.Helper()

	googleMock.RegisterCode(
		code,
		mocks.GoogleUserInfo{
			ID:            "google_user_123",
			Email:         email,
			VerifiedEmail: true,
			Name:          name,
			Picture:       "https://example.com/pic.jpg",
		},
	)

	req := newExchangeCodeRequest(t, authv1.Provider_PROVIDER_GOOGLE, code)
	appCreds, ok := st.GetAppCredentials("auth_client")
	require.True(t, ok, "auth_client credentials must exist in config")
	req.Header().Set("Authorization", "Bearer "+appCreds.BearerToken)

	resp, err := st.AuthClient.ExchangeCode(ctx, req)
	require.NoError(t, err)

	return resp.Msg.AccessToken, resp.Msg.RefreshToken
}
