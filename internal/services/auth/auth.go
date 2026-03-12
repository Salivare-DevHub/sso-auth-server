package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/salivare-io/slogx"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
	userdom "github.com/salivare-io/sso-auth-server/internal/domain/user"
	authjwt "github.com/salivare-io/sso-auth-server/internal/lib/jwt"
)

// ErrUnknownProvider is returned when provider is not supported.
var ErrUnknownProvider = errors.New("unknown provider")
var ErrInvalidAccessToken = errors.New("invalid access token")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

// UserGetter gets a user by ID or email.
type UserGetter interface {
	GetByID(ctx context.Context, userID string) (*userdom.User, error)
	GetByEmail(ctx context.Context, email string) (*userdom.User, error)
}

// UserSaver saves or creates a user.
type UserSaver interface {
	SaveUser(ctx context.Context, email, name string) (string, error)
}

type IdentitySource interface {
	FetchUser(ctx context.Context, code string) (*models.User, error)
}

// RefreshStore stores refresh tokens.
type RefreshStore interface {
	Save(ctx context.Context, userID, sessionID, refresh, appID string) error
	GetByRefreshToken(ctx context.Context, refresh string) (*models.SessionData, error)
	DeleteByRefreshToken(ctx context.Context, refresh string) error
}

// AppProvider provides app lookup by ID.
type AppProvider interface {
	GetByID(ctx context.Context, id string) (*models.App, error)
}

// PermissionsStore provides user permissions.
type PermissionsStore interface {
	GetByUserID(ctx context.Context, userID string) ([]models.Permission, error)
}

// Service provides auth-related operations.
type Service struct {
	log         *slogx.Logger
	providers   map[string]IdentitySource
	userSaver   UserSaver
	userGetter  UserGetter
	refresh     RefreshStore
	appProvider AppProvider
	permissions PermissionsStore
}

// New creates a new Service.
func New(
	log *slogx.Logger,
	providers map[string]IdentitySource,
	userSaver UserSaver,
	userGetter UserGetter,
	refresh RefreshStore,
	appProvider AppProvider,
	permissions PermissionsStore,
) *Service {
	return &Service{
		log:         log,
		providers:   providers,
		userSaver:   userSaver,
		userGetter:  userGetter,
		refresh:     refresh,
		appProvider: appProvider,
		permissions: permissions,
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

	userID, err := s.userSaver.SaveUser(ctx, userInfo.Email, userInfo.Name)
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

	access, err := authjwt.NewAccessToken(userID, sessionID.String(), *app)
	if err != nil {
		log.Error("failed to generate access token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate access token: %w", op, err)
	}

	refreshToken, err := authjwt.NewRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: generate refresh token: %w", op, err)
	}

	if err := s.refresh.Save(ctx, userID, sessionID.String(), refreshToken, appID); err != nil {
		log.Error("failed to save refresh token", slog.String("err", err.Error()))
		return "", "", fmt.Errorf("%s: save refresh token: %w", op, err)
	}

	log.Info(
		"login completed successfully", slog.String("user_id", userID), slog.String("session_id", sessionID.String()),
	)

	return access, refreshToken, nil
}

// UserDetails returns user info for a valid access token.
func (s *Service) UserDetails(ctx context.Context, accessToken string) (*models.User, error) {
	if accessToken == "" {
		return nil, ErrInvalidAccessToken
	}

	claims, err := s.parseAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userGetter.GetByID(ctx, claims.UID)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

// Refresh rotates refresh token and returns new access/refresh tokens.
func (s *Service) Refresh(ctx context.Context, appID, refreshToken string) (string, string, error) {
	if refreshToken == "" {
		return "", "", ErrInvalidRefreshToken
	}

	session, err := s.refresh.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}
	if session.AppID != appID {
		return "", "", ErrInvalidRefreshToken
	}

	if err := s.refresh.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return "", "", ErrInvalidRefreshToken
	}

	newSessionID, err := uuid.NewV7()
	if err != nil {
		return "", "", fmt.Errorf("generate session id: %w", err)
	}

	app, err := s.appProvider.GetByID(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("get app: %w", err)
	}

	access, err := authjwt.NewAccessToken(session.UserID, newSessionID.String(), *app)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := authjwt.NewRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.refresh.Save(ctx, session.UserID, newSessionID.String(), newRefreshToken, appID); err != nil {
		return "", "", fmt.Errorf("save refresh token: %w", err)
	}

	return access, newRefreshToken, nil
}

// Logout invalidates a refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidRefreshToken
	}

	if err := s.refresh.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return ErrInvalidRefreshToken
	}

	return nil
}

// RevokeRefreshToken invalidates a refresh token.
func (s *Service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return s.Logout(ctx, refreshToken)
}

// GetPermissions returns permissions for a valid access token.
func (s *Service) GetPermissions(ctx context.Context, accessToken string) ([]models.Permission, error) {
	if accessToken == "" {
		return nil, ErrInvalidAccessToken
	}

	claims, err := s.parseAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return s.permissions.GetByUserID(ctx, claims.UID)
}

func (s *Service) parseAccessToken(ctx context.Context, accessToken string) (*authjwt.Claims, error) {
	parser := jwtv5.NewParser(jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}))

	claims := &authjwt.Claims{}
	_, _, err := parser.ParseUnverified(accessToken, claims)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}

	if claims.AppID == "" {
		return nil, ErrInvalidAccessToken
	}

	app, err := s.appProvider.GetByID(ctx, claims.AppID)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}
	if app.SigningKey == "" {
		return nil, ErrInvalidAccessToken
	}

	token, err := jwtv5.ParseWithClaims(
		accessToken,
		claims,
		func(token *jwtv5.Token) (interface{}, error) {
			return []byte(app.SigningKey), nil
		},
		jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidAccessToken
	}

	return claims, nil
}
