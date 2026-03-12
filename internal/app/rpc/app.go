package rpcapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/grpcreflect"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	authconnect "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1/authservicev1connect"
	"github.com/salivare-io/slogx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/salivare-io/sso-auth-server/internal/config"
	appmanager "github.com/salivare-io/sso-auth-server/internal/domain/app"
	rpcmiddleware "github.com/salivare-io/sso-auth-server/internal/middleware/rpc"
	authrpc "github.com/salivare-io/sso-auth-server/internal/rpchttp/auth"
)

// App is the RPC HTTP server wrapper.
type App struct {
	log             *slogx.Logger
	router          chi.Router
	server          *http.Server
	host            string
	port            int
	shutdownTimeout time.Duration
}

// New creates a new RPC HTTP server.
func New(
	log *slogx.Logger,
	authService authrpc.Auth,
	appManager *appmanager.Manager,
	httpCfg config.HTTPConfig,
	envCfh string,
) *App {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(rpcmiddleware.LoggerContext(log))
	r.Use(rpcmiddleware.Logger(log))
	r.Use(middleware.Timeout(httpCfg.Timeout))

	// Use Bearer middleware with AppManager to validate bearer tokens.
	r.Use(rpcmiddleware.Bearer(appManager))

	// mount generated Connect handler via adapter register
	authrpc.Register(r, authService)

	enableReflection := envCfh == config.EnvLocal || envCfh == config.EnvDev
	registerReflection(r, log, enableReflection)

	addr := net.JoinHostPort(httpCfg.Host, strconv.Itoa(httpCfg.Port))

	server := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(r, &http2.Server{}),
		ReadTimeout:  httpCfg.Timeout,
		WriteTimeout: httpCfg.Timeout,
	}

	return &App{
		log:             log,
		router:          r,
		server:          server,
		host:            httpCfg.Host,
		port:            httpCfg.Port,
		shutdownTimeout: httpCfg.ShutdownTimeout,
	}
}

// MustRun starts the server or panics on error.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run starts serving HTTP requests.
func (a *App) Run() error {
	const op = "rpcapp.Run"

	ln, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.Int("port", a.port),
	)

	log.Info("RPC HTTP server is running", slog.String("address", a.server.Addr))

	if err := a.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Stop gracefully shuts down the server.
func (a *App) Stop() {
	const op = "rpcapp.Stop"

	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.Int("port", a.port),
	)

	log.Info("RPC HTTP server is stopping")

	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Warn("graceful shutdown failed, forcing close", slog.Any("err", err))
		_ = a.server.Close()
		return
	}

	log.Info("RPC HTTP server stopped gracefully")
}

func registerReflection(r chi.Router, log *slogx.Logger, enable bool) {
	if !enable {
		return
	}

	reflector := grpcreflect.NewStaticReflector(
		authconnect.AuthServiceName,
	)

	// reflection handlers
	pathV1, handlerV1 := grpcreflect.NewHandlerV1(reflector)
	pathV1alpha, handlerV1alpha := grpcreflect.NewHandlerV1Alpha(reflector)

	// mount reflection handlers at returned paths (root)
	r.Handle(pathV1, handlerV1)
	r.Handle(pathV1alpha, handlerV1alpha)

	log.Info("registered gRPC reflection", slog.String("path_v1", pathV1))
}
