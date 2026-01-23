package app

import (
	"fmt"
	grpcapp "github.com/Salivare-DevHub/sso-auth-server/internal/app/grpc"
	"github.com/Salivare-DevHub/sso-auth-server/internal/config"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/auth/providers"
	"github.com/Salivare-DevHub/sso-auth-server/internal/grpc/user"
	"github.com/Salivare-DevHub/sso-auth-server/internal/providers/appconfig"
	"github.com/Salivare-DevHub/sso-auth-server/internal/providers/oauth"
	"github.com/Salivare-DevHub/sso-auth-server/internal/services/auth"
	"github.com/Salivare-DevHub/sso-auth-server/internal/storage/redis"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	AuthConfig config.AuthConfig,
) (*App, error) {
	googleOAuth := oauth.NewGoogle(
		AuthConfig.Google.ClientID,
		AuthConfig.Google.ClientSecret,
		AuthConfig.Google.RedirectURL,
	)

	yandexOAuth := oauth.NewYandex(
		AuthConfig.Yandex.ClientID,
		AuthConfig.Yandex.ClientSecret,
	)

	identitySource := map[string]auth.IdentitySource{
		"google": providers.NewGoogleIdentity(googleOAuth),
		"yandex": providers.NewYandexIdentity(yandexOAuth),
	}

	appProvider, err := appconfig.NewYAML("./configs/appslocal.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	userStore := user.NewClient()
	redisClient := redis.New("localhost:6379")
	refreshStore := redis.NewRefreshStore(redisClient, time.Hour*24*30)

	authService := auth.New(log, identitySource, userStore, refreshStore, appProvider)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}, nil
}
