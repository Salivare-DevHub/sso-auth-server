package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"log/slog"
)

type UserStore interface {
	FindOrCreate(ctx context.Context, email string) (string, error)
}

type RefreshStore interface {
	Save(ctx context.Context, userID string, refresh string) error
}

type TokenService interface {
	Generate(userID string) (access string, refresh string)
}

type Provider interface {
	AuthURL(ctx context.Context) string
	Exchange(ctx context.Context, code string) (*models.User, error)
}

const (
	googleProvider = "google"
	yandexProvider = "yandex"
)

var (
	ErrUnknownProvider = errors.New("unknown provider")
	ErrExchangeFailed  = errors.New("failed to exchange oauth code")
)

type Providers struct {
	Google Provider
	Yandex Provider
}

type Service struct {
	log       *slog.Logger
	providers Providers
	users     UserStore
	refresh   RefreshStore
	tokens    TokenService
}

func New(
	log *slog.Logger,
	google Provider,
	yandex Provider,
	users UserStore,
	refresh RefreshStore,
	tokens TokenService,
) *Service {
	return &Service{
		log: log,
		providers: Providers{
			Google: google,
			Yandex: yandex,
		},
		users:   users,
		refresh: refresh,
		tokens:  tokens,
	}
}

// StartAuth initiates an OAuth2 authorization process for the specified provider.
// The method selects the appropriate OAuth provider (Google or Yandex),
// generates a redirect-URL and returns it to the caller.
// Returns an ErrUnknownProvider error if the provider is not supported.
func (s *Service) StartAuth(ctx context.Context, providerName string) (string, error) {
	const op = "auth.StartAuth"

	log := s.log.With(
		slog.String("op", op),
		slog.String("provider", providerName),
	)

	p, err := s.getProvider(providerName)
	if err != nil {
		log.Warn("unknown provider")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	url := p.AuthURL(ctx)

	log.Info("generated auth redirect url")

	return url, nil
}

// ExchangeCode completes OAuth2 authorization by exchanging the received authorization code
// per user profile at the selected provider (Google or Yandex).
// After receiving the user’s data, the service creates or finds an entry in the repository,
// generates its own access and refresh tokens and saves the refresh token.
// Can return ErrUnknownProvider or ErrExchangeFailed.
func (s *Service) ExchangeCode(ctx context.Context, providerName string, code string) (string, string, error) {
	const op = "auth.ExchangeCode"

	log := s.log.With(
		slog.String("op", op),
		slog.String("provider", providerName),
	)

	p, err := s.getProvider(providerName)
	if err != nil {
		log.Warn("unknown provider")
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	userInfo, err := p.Exchange(ctx, code)
	if err != nil {
		log.Error("failed to exchange oauth code", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, ErrExchangeFailed)
	}

	userID, err := s.users.FindOrCreate(ctx, userInfo.Email)
	if err != nil {
		log.Error("failed to find or create user", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	access, refresh := s.tokens.Generate(userID)

	if err := s.refresh.Save(ctx, userID, refresh); err != nil {
		log.Error("failed to save refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("oauth exchange completed successfully")

	return access, refresh, nil
}

// getProvider returns the OAuth provider by its name.
// Used to select a specific implementation (Google or Yandex)
// based on a string obtained from the outer layer (for example, gRPC).
// If no provider is found, returns ErrUnknownProvider and nil as the provider.
func (s *Service) getProvider(providerName string) (Provider, error) {
	switch providerName {
	case googleProvider:
		return s.providers.Google, nil
	case yandexProvider:
		return s.providers.Yandex, nil
	default:
		return nil, ErrUnknownProvider
	}
}
