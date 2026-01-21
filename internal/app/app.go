package app

import (
	grpcapp "github.com/Salivare-DevHub/sso-auth-server/internal/app/grpc"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/auth/providers"
	"github.com/Salivare-DevHub/sso-auth-server/internal/provider/oauth"
	"github.com/Salivare-DevHub/sso-auth-server/internal/services/auth"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
) *App {
	googleOAuth := oauth.NewGoogleOAuth("id", "secret")
	yandexOAuth := oauth.NewYandexOAuth("id", "secret")

	authStrategy := map[string]auth.LoginStrategy{
		"google": providers.NewOAuthLogin(googleOAuth, users, refresh),
		"yandex": providers.NewOAuthLogin(yandexOAuth, users, refresh),
	}

	authService := auth.New(log, authStrategy)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
