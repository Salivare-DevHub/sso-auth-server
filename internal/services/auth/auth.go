package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"github.com/Salivare-DevHub/sso-auth-server/internal/lib/jwt"
)

var (
	ErrUnknownProvider = errors.New("unknown provider")
)

type IdentitySource interface {
	FetchUser(ctx context.Context, code string) (*models.User, error)
}

type UserStore interface {
	FindOrCreate(ctx context.Context, email string) (string, error)
}

type RefreshStore interface {
	Save(ctx context.Context, userID string, refresh string) error
}

type AppProvider interface {
	GetByID(ctx context.Context, id int64) (*models.App, error)
}

type Service struct {
	log       *slog.Logger
	providers map[string]IdentitySource
	users     UserStore
	refresh   RefreshStore
	apps      AppProvider
}

func New(
	log *slog.Logger,
	providers map[string]IdentitySource,
	users UserStore,
	refresh RefreshStore,
	apps AppProvider,
) *Service {
	return &Service{
		log:       log,
		providers: providers,
		users:     users,
		refresh:   refresh,
		apps:      apps,
	}
}

func (s *Service) Login(ctx context.Context, provider string, code string, appID int64) (string, string, error) {
	const op = "auth.Service.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("provider", provider),
	)

	src, ok := s.providers[provider]
	if !ok {
		return "", "", ErrUnknownProvider
	}

	userInfo, err := src.FetchUser(ctx, code)
	if err != nil {
		return "", "", fmt.Errorf("%s: fetch user: %w", op, err)
	}

	userID, err := s.users.FindOrCreate(ctx, userInfo.Email)
	if err != nil {
		return "", "", fmt.Errorf("%s: find or create user: %w", op, err)
	}

	app, err := s.apps.GetByID(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("%s: get app: %w", op, err)
	}

	access, err := jwt.NewAccessToken(userID, *app)
	if err != nil {
		return "", "", fmt.Errorf("%s: generate access token: %w", op, err)
	}

	refreshToken, err := jwt.NewRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("%s: generate refresh token: %w", op, err)
	}

	if err := s.refresh.Save(ctx, userID, refreshToken); err != nil {
		return "", "", fmt.Errorf("%s: save refresh token: %w", op, err)
	}

	log.Info("login completed successfully")

	return access, refreshToken, nil
}
