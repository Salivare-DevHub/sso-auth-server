package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/salivare-io/slogx"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
	userdom "github.com/salivare-io/sso-auth-server/internal/domain/user"
	"github.com/salivare-io/sso-auth-server/internal/lib/jwt"
)

// ErrUnknownProvider is returned when provider is not supported.
var ErrUnknownProvider = errors.New("unknown provider")

// UserGetter gets a user by ID or email.
type UserGetter interface {
	GetByID(ctx context.Context, userID string) (*userdom.User, error)
	GetByEmail(ctx context.Context, email string) (*userdom.User, error)
}

// UserSaver saves or creates a user.
type UserSaver interface {
	SaveUser(ctx context.Context, email string) (string, error)
}

type IdentitySource interface {
	FetchUser(ctx context.Context, code string) (*models.User, error)
}

// RefreshStore stores refresh tokens.
type RefreshStore interface {
	Save(ctx context.Context, userID string, sessionID string, refresh string) error
}

// AppProvider provides app lookup by ID.
type AppProvider interface {
	GetByID(ctx context.Context, id string) (*models.App, error)
}

// Service provides auth-related operations.
type Service struct {
	log         *slogx.Logger
	providers   map[string]IdentitySource
	userSaver   UserSaver
	refresh     RefreshStore
	appProvider AppProvider
}

// New creates a new Service.
func New(
	log *slogx.Logger,
	providers map[string]IdentitySource,
	userSaver UserSaver,
	refresh RefreshStore,
	appProvider AppProvider,
) *Service {
	return &Service{
		log:         log,
		providers:   providers,
		userSaver:   userSaver,
		refresh:     refresh,
		appProvider: appProvider,
	}
}

// Login exchanges OAuth code and returns access/refresh tokens.
func (s *Service) Login(ctx context.Context, provider string, code string, appID string) (string, string, error) {
	const op = "auth.Service.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("provider", provider),
		slog.String("app_id", appID),
	)

	src, ok := s.providers[provider]
	if !ok {
		log.Error("failed to find provider")
		return "", "", ErrUnknownProvider
	}

	userInfo, err := src.FetchUser(ctx, code)
	if err != nil {
		log.Error("failed to fetch user", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: fetch user: %w", op, err)
	}

	userID, err := s.userSaver.SaveUser(ctx, userInfo.Email)
	if err != nil {
		log.Error("failed to save user", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: find or create user: %w", op, err)
	}

	// Generate sessionID for this session.
	sessionID, err := uuid.NewV7()
	if err != nil {
		log.Error("failed to generate session id", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate session id: %w", op, err)
	}

	// Fetch the app from the provider.
	app, err := s.appProvider.GetByID(ctx, appID)
	if err != nil {
		log.Error("failed to get app", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: get app: %w", op, err)
	}

	access, err := jwt.NewAccessToken(userID, sessionID.String(), *app)
	if err != nil {
		log.Error("failed to generate access token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate access token: %w", op, err)
	}

	refreshToken, err := jwt.NewRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate refresh token: %w", op, err)
	}

	if err := s.refresh.Save(ctx, userID, sessionID.String(), refreshToken); err != nil {
		log.Error("failed to save refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: save refresh token: %w", op, err)
	}

	log.Info(
		"login completed successfully", slog.String("user_id", userID), slog.String("session_id", sessionID.String()),
	)

	return access, refreshToken, nil
}
