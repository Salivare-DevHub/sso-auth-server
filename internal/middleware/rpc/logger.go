package rpcmiddleware

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/salivare-io/slogx"
)

const LogFieldRequestID = "request_id"

// GetRequestID возвращает request id из контекста, который ставит chi middleware.RequestID.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(chimw.RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// LoggerContext кладёт контекстный slogx.Logger в контекст запроса.
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

// Logger логирует метод, путь и длительность запроса.
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
