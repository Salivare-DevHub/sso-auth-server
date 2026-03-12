package rpcmiddleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	connect "connectrpc.com/connect"

	"github.com/salivare-io/sso-auth-server/internal/domain/app"
)

type (
	ctxKeyAuth  struct{}
	ctxKeyAppID struct{}
)

type AuthInfo struct{} // AuthInfo is a marker type for auth context.

var errUnauthorized = errors.New("unauthorized")

// Bearer returns middleware that validates Authorization: Bearer <token> via AppManager.
func Bearer(appManager *app.Manager) func(http.Handler) http.Handler {
	if appManager == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				errWriter := connect.NewErrorWriter()

				h := r.Header.Get("Authorization")
				parts := strings.Fields(h)
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					err := connect.NewError(connect.CodeUnauthenticated, errUnauthorized)
					if writeErr := errWriter.Write(w, r, err); writeErr != nil {
						http.Error(w, errUnauthorized.Error(), http.StatusUnauthorized)
					}
					return
				}

				token := parts[1]

				// Validate the token via AppManager.
				appID, err := appManager.VerifyBearerToken(r.Context(), token)
				if err != nil {
					err := connect.NewError(connect.CodeUnauthenticated, errUnauthorized)
					if writeErr := errWriter.Write(w, r, err); writeErr != nil {
						http.Error(w, errUnauthorized.Error(), http.StatusUnauthorized)
					}
					return
				}

				// Attach auth info and app_id to the context.
				ctx := context.WithValue(r.Context(), ctxKeyAuth{}, &AuthInfo{})
				ctx = context.WithValue(ctx, ctxKeyAppID{}, appID)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

// FromContext extracts the auth marker from the context.
func FromContext(ctx context.Context) (*AuthInfo, bool) {
	v := ctx.Value(ctxKeyAuth{})
	if v == nil {
		return nil, false
	}
	ai, ok := v.(*AuthInfo)
	return ai, ok
}

// GetAppIDFromContext extracts app_id from the context.
func GetAppIDFromContext(ctx context.Context) string {
	appID, ok := ctx.Value(ctxKeyAppID{}).(string)
	if !ok {
		return ""
	}
	return appID
}
