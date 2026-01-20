package yandex

import (
	"context"
	"fmt"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"github.com/Salivare-DevHub/sso-auth-server/internal/services/auth"
	"golang.org/x/oauth2"
)

type Auth struct {
	config *oauth2.Config
}

var _ auth.Provider = (*Auth)(nil)

func New(clientID, clientSecret, redirectURL string) *Auth {
	return &Auth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"login:email", "login:info"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://oauth.yandex.ru/authorize",
				TokenURL: "https://oauth.yandex.ru/token",
			},
		},
	}
}

func (y *Auth) AuthURL(ctx context.Context) string {
	return y.config.AuthCodeURL("state")
}

func (y *Auth) Exchange(ctx context.Context, code string) (*models.User, error) {
	token, err := y.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("yandex exchange failed: %w", err)
	}

	// TODO: заменить заглушку
	return &models.User{
		Email: "example@yandex.ru",
		Name:  "Yandex User",
		ID:    token.AccessToken,
	}, nil
}
