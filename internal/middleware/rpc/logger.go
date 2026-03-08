package rpcmiddleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/salivare-io/slogx"
)

const LogFieldRequestID = "request_id"

// GetRequestID returns a request id from the context that sets chi middleware.RequestID.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(chimw.RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// LoggerContext puts a context slogx. Logger in the query context.
func LoggerContext(log *slogx.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				reqID := GetRequestID(r.Context())
				logger := log.With(slog.String(LogFieldRequestID, reqID))
				ctx := slogx.ToContext(r.Context(), logger)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

// Logger Logins the method, path, and duration of the request.
func Logger(log *slogx.Logger) func(http.Handler) http.Handler {
	log = log.With(slog.String("component", "http"))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				next.ServeHTTP(w, r)
				log.Info(
					"request completed",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.String(LogFieldRequestID, GetRequestID(r.Context())),
					slog.Duration("duration", time.Since(start)),
				)
			},
		)
	}
}
