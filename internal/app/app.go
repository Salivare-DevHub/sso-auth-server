package app

import (
	grpcapp "github.com/Salivare-DevHub/sso-auth-server/internal/app/grpc"
	gprovider "github.com/Salivare-DevHub/sso-auth-server/internal/provider/google"
	yprovider "github.com/Salivare-DevHub/sso-auth-server/internal/provider/yandex"
	"github.com/Salivare-DevHub/sso-auth-server/internal/services/auth"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	tokenTTL time.Duration,
) *App {
	// TODO: init storage
	// TODO: Remove hardcode
	googleProvider := gprovider.New(
		"google-client-id",
		"google-secret",
		"http://localhost:8080/auth/google/callback",
	)
	yandexProvider := yprovider.New(
		"yandex-client-id",
		"yandex-secret",
		"http://localhost:8080/auth/yandex/callback",
	)

	// TODO: Добавить users, refresh, tokens
	authService := auth.New(
		log,
		googleProvider,
		yandexProvider,
		users,
		refresh,
		tokens,
	)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
