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
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {

	identitySource := make(map[string]auth.IdentitySource)

	for name, provider := range cfg.OAuthProviders {
		switch name {
		case providers.ProviderGoogle:
			identitySource[name] = providers.NewGoogleIdentity(
				oauth.NewGoogle(provider.ClientID, provider.ClientSecret, provider.RedirectURL),
			)
		case providers.ProviderYandex:
			identitySource[name] = providers.NewYandexIdentity(
				oauth.NewYandex(provider.ClientID, provider.ClientSecret),
			)
		}
	}

	appProvider := appconfig.NewInternalApps(cfg.InternalApps)

	userStore := user.NewClient()
	redisClient := redis.New(fmt.Sprintf(":%d", cfg.Redis.Port))
	refreshStore := redis.NewRefreshStore(redisClient, cfg.TokenTTL)

	authService := auth.New(log, identitySource, userStore, refreshStore, appProvider)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC)

	return &App{
		GRPCSrv: grpcApp,
	}, nil
}
