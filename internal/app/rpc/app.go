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
	authconnect "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1/authservicev1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/Salivare-DevHub/sso-auth-server/internal/config"
	authrpc "github.com/Salivare-DevHub/sso-auth-server/internal/rpc/auth"
)

type App struct {
	log             *slog.Logger
	router          chi.Router
	server          *http.Server
	host            string
	port            int
	shutdownTimeout time.Duration
}

func New(
	log *slog.Logger,
	authService authrpc.Auth,
	httpCfg config.HTTPConfig,
	envCfh string,
) *App {
	r := chi.NewRouter()

	// mount generated Connect handler via adapter register
	authrpc.Register(r, authService)

	if envCfh == config.EnvLocal || envCfh == config.EnvProd {
		reflector := grpcreflect.NewStaticReflector(
			authconnect.AuthServiceName,
		)

		// reflection handlers
		pathV1, handlerV1 := grpcreflect.NewHandlerV1(reflector)
		pathV1alpha, handlerV1alpha := grpcreflect.NewHandlerV1Alpha(reflector)

		// mount reflection handlers at returned paths (root)
		r.Handle(pathV1, handlerV1)
		r.Handle(pathV1alpha, handlerV1alpha)

		log.Info("Registered with grpc reflection")
	}

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

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "rpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.Int("port", a.port),
	)

	log.Info("RPC HTTP server is running", slog.String("address", a.server.Addr))

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

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
