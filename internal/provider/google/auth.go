package google

import (
	"context"
	"fmt"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"github.com/Salivare-DevHub/sso-auth-server/internal/services/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *Auth) AuthURL(ctx context.Context) string {
	return g.config.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (g *Auth) Exchange(ctx context.Context, code string) (*models.User, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google exchange failed: %w", err)
	}

	// TODO: заменить заглушку
	return &models.User{
		Email: "example@gmail.com",
		Name:  "Google User",
		ID:    token.AccessToken,
	}, nil
}
