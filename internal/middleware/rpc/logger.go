package rpcmiddleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/salivare-io/slogx"
)

// LogFieldRequestID is the log field name for request IDs.
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

	return chimw.RequestLogger(&slogFormatter{log: log})
}

type slogFormatter struct {
	log *slogx.Logger
}

func (f *slogFormatter) NewLogEntry(r *http.Request) chimw.LogEntry {
	entry := f.log.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.String(LogFieldRequestID, GetRequestID(r.Context())),
	)

	return &slogEntry{log: entry}
}

type slogEntry struct {
	log *slogx.Logger
}

func (e *slogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	e.log.Info(
		"request completed",
		slog.Int("status", status),
		slog.Int("bytes", bytes),
		slog.Duration("duration", elapsed),
	)
}

func (e *slogEntry) Panic(v interface{}, stack []byte) {
	e.log.Error(
		"request panic",
		slog.Any("panic", v),
		slog.String("stack", string(stack)),
	)
}
