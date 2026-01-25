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
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	host       string
	port       int
}

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	cfg config.GRPCConfig,
) *App {
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmw.Timeout(cfg.Timeout)),
	)

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		host:       cfg.Host,
		port:       cfg.Port,
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

	a.gRPCServer.GracefulStop()
}
