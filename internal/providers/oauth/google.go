package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleUserInfoUrl = "https://openidconnect.googleapis.com/v1/userinfo"
	ScopeOpenID       = "openid"
	ScopeEmail        = "https://www.googleapis.com/auth/userinfo.email"
	ScopeProfile      = "https://www.googleapis.com/auth/userinfo.profile"
)

type Google struct {
	config *oauth2.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewGoogle(clientID, clientSecret, redirectURL string) *Google {
	return &Google{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				ScopeOpenID,
				ScopeProfile,
				ScopeEmail,
			},
		},
	}
}

func (g *Google) Exchange(ctx context.Context, code string) (*models.User, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google exchange failed: %w", err)
	}

	client := g.config.Client(ctx, token)

	resp, err := client.Get(googleUserInfoUrl)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google api returned status: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &models.User{
		Email: userInfo.Email,
		Name:  userInfo.Name,
		ID:    userInfo.ID,
	}, nil
}
