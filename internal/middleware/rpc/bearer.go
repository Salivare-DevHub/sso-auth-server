package rpcmiddleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
)

type ctxKeyAuth struct{}

type AuthInfo struct{} // пустая структура — маркер авторизации

// Bearer возвращает middleware, который сравнивает Authorization: Bearer <token>
func Bearer(expectedToken string) func(http.Handler) http.Handler {
	if expectedToken == "" {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				h := r.Header.Get("Authorization")
				parts := strings.Fields(h)
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedToken)) != 1 {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), ctxKeyAuth{}, &AuthInfo{})
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

// FromContext извлекает маркер авторизации из контекста
func FromContext(ctx context.Context) (*AuthInfo, bool) {
	v := ctx.Value(ctxKeyAuth{})
	if v == nil {
		return nil, false
	}
	ai, ok := v.(*AuthInfo)
	return ai, ok
}
