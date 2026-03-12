package app

import (
	"net"
	"strconv"

	"github.com/salivare-io/slogx"

	rpcapp "github.com/salivare-io/sso-auth-server/internal/app/rpc"
	"github.com/salivare-io/sso-auth-server/internal/config"
	appmanager "github.com/salivare-io/sso-auth-server/internal/domain/app"
	"github.com/salivare-io/sso-auth-server/internal/domain/auth/providers"
	"github.com/salivare-io/sso-auth-server/internal/domain/models"
	userrepo "github.com/salivare-io/sso-auth-server/internal/domain/user"
	"github.com/salivare-io/sso-auth-server/internal/providers/oauth"
	"github.com/salivare-io/sso-auth-server/internal/services/auth"
	"github.com/salivare-io/sso-auth-server/internal/storage/permissions"
	"github.com/salivare-io/sso-auth-server/internal/storage/redis"
)

// App wires together RPC server and dependencies.
type App struct {
	RPCSrv *rpcapp.App
}

var identitySourceBuilders = map[string]func(config.OAuthProvider) auth.IdentitySource{
	providers.ProviderGoogle: func(p config.OAuthProvider) auth.IdentitySource {
		return providers.NewGoogleIdentity(
			oauth.NewGoogle(
				p.ClientID,
				p.ClientSecret,
				p.RedirectURL,
				p.TokenURL,
				p.UserInfoURL,
				p.Scopes,
			),
		)
	},
	providers.ProviderYandex: func(p config.OAuthProvider) auth.IdentitySource {
		return providers.NewYandexIdentity(
			oauth.NewYandex(
				p.ClientID,
				p.ClientSecret,
				p.RedirectURL,
				p.TokenURL,
				p.UserInfoURL,
				p.Scopes,
			),
		)
	},
}

// New builds the application with all dependencies.
func New(log *slogx.Logger, cfg *config.Config) (*App, error) {
	identitySource := make(map[string]auth.IdentitySource)

	for name, provider := range cfg.OAuthProviders {
		if build, ok := identitySourceBuilders[name]; ok {
			identitySource[name] = build(provider)
		}
	}

	userStore := userrepo.NewInMemoryRepository()

	redisClient := redis.New(net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port)))
	refreshStore := redis.NewRefreshStore(redisClient, cfg.TokenTTL)

	// Initialize AppManager for apps and bearer token management.
	appsMap := convertInternalAppsToModels(cfg.InternalApps)
	appManager := appmanager.NewAppManager(appsMap)

	permissionsStore := permissions.NewInMemoryStore()

	authService := auth.New(log, identitySource, userStore, userStore, refreshStore, appManager, permissionsStore)

	rpcApp := rpcapp.New(log, authService, appManager, cfg.HTTP, cfg.Env)

	return &App{
		RPCSrv: rpcApp,
	}, nil
}

// convertInternalAppsToModels converts config to app models.
func convertInternalAppsToModels(appsConfig map[string]config.AppCredentials) map[string]*models.App {
	appsMap := make(map[string]*models.App)
	for appID, cred := range appsConfig {
		appsMap[appID] = &models.App{
			ID:          appID,
			Name:        cred.ClientID,
			BearerToken: cred.BearerToken,
			SigningKey:  cred.SigningKey,
			AccessTTL:   cred.AccessTTL,
			RefreshTTL:  cred.RefreshTTL,
		}
	}
	return appsMap
}
