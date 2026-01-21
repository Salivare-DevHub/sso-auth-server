package providers

import (
	"context"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"github.com/Salivare-DevHub/sso-auth-server/internal/lib/jwt"
)

type OAuthProvider interface {
	Exchange(ctx context.Context, code string) (*models.User, error)
}

type UserStore interface {
	FindOrCreate(ctx context.Context, email string) (string, error)
}

type RefreshStore interface {
	Save(ctx context.Context, userID string, refresh string) error
}

type OAuthLogin struct {
	oauth   OAuthProvider
	users   UserStore
	refresh RefreshStore
}

func NewOAuthLogin(
	oauth OAuthProvider,
	users UserStore,
	refresh RefreshStore,
) *OAuthLogin {
	return &OAuthLogin{
		oauth:   oauth,
		users:   users,
		refresh: refresh,
	}
}

func (s *OAuthLogin) Login(ctx context.Context, code string) (string, string, error) {
	userInfo, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return "", "", err
	}

	userID, err := s.users.FindOrCreate(ctx, userInfo.Email)
	if err != nil {
		return "", "", err
	}

	access, err := jwt.NewAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refresh, err := jwt.NewRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := s.refresh.Save(ctx, userID, refresh); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}
