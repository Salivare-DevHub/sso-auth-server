package auth

import (
	"context"
	"errors"
	"fmt"
	grpcmw "github.com/Salivare-DevHub/sso-auth-server/internal/middleware/grpc"
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

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
	) (uid int64, err error)
}

type RefreshStore interface {
	Save(ctx context.Context, userID int64, refresh string) error
}

type AppProvider interface {
	GetByID(ctx context.Context, id int64) (*models.App, error)
}

type Service struct {
	log       *slog.Logger
	providers map[string]IdentitySource
	userSaver UserSaver
	refresh   RefreshStore
}

func New(
	log *slog.Logger,
	providers map[string]IdentitySource,
	userSaver UserSaver,
	refresh RefreshStore,
) *Service {
	return &Service{
		log:       log,
		providers: providers,
		userSaver: userSaver,
		refresh:   refresh,
	}
}

func (s *Service) Login(ctx context.Context, provider string, code string) (string, string, error) {
	const op = "auth.Service.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("provider", provider),
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

	appID, err := grpcmw.GetAppID(ctx)
	if err != nil {
		log.Error("failed to get app_id", slog.String("err", err.Error()))
		return "", "", err
	}

	appName, err := grpcmw.GetAppName(ctx)
	if err != nil {
		log.Error("failed to get app_name", slog.String("err", err.Error()))
		return "", "", err
	}

	app := &models.App{
		ID:   appID,
		Name: appName,
	}

	access, err := jwt.NewAccessToken(userID, *app)
	if err != nil {
		log.Error("failed to generate access token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate access token: %w", op, err)
	}

	refreshToken, err := jwt.NewRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate refresh token: %w", op, err)
	}

	if err := s.refresh.Save(ctx, userID, refreshToken); err != nil {
		log.Error("failed to save refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: save refresh token: %w", op, err)
	}

	log.Info("login completed successfully")

	return access, refreshToken, nil
}
