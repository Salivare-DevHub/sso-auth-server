package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireAppCredentials gets app credentials from Suite and fails the test if missing.
func RequireAppCredentials(t *testing.T, appName string, getCreds func(string) (interface{}, bool)) string {
	t.Helper()

	creds, ok := getCreds(appName)
	require.True(t, ok, "credentials for app %q must exist in config", appName)

	if bearerToken, ok := creds.(string); ok {
		return bearerToken
	}

	require.Fail(t, "credentials for app %q is not a string", appName)
	return ""
}
