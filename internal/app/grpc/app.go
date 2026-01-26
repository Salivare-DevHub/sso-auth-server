package grpcapp

import (
	"fmt"
	"github.com/Salivare-DevHub/sso-auth-server/internal/config"
	authgrpc "github.com/Salivare-DevHub/sso-auth-server/internal/grpc/auth"
	grpcmw "github.com/Salivare-DevHub/sso-auth-server/internal/middleware/grpc"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"strconv"
	"time"
)

type App struct {
	log             *slog.Logger
	gRPCServer      *grpc.Server
	host            string
	port            int
	shutdownTimeout time.Duration
}

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	grpcCfg config.GRPCConfig,
	internalApps map[string]config.AppCredentials,
) *App {
	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcmw.InternalAuth(internalApps),
			grpcmw.Timeout(grpcCfg.Timeout),
		),
	)

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:             log,
		gRPCServer:      gRPCServer,
		host:            grpcCfg.Host,
		port:            grpcCfg.Port,
		shutdownTimeout: grpcCfg.ShutdownTimeout,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.Int("port", a.port),
	)

	addr := net.JoinHostPort(a.host, strconv.Itoa(a.port))
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("address", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	log := a.log.With(slog.String("op", op))
	log.Info("gRPC server is stopping", slog.Int("port", a.port))

	done := make(chan struct{})

	go func() {
		a.gRPCServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("gRPC server stopped gracefully")
	case <-time.After(a.shutdownTimeout):
		log.Warn("Graceful stop timed out, forcing stop")
		a.gRPCServer.Stop()
	}
}
