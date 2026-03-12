package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
	"golang.org/x/oauth2"
)

// Google implements OAuth for Google.
type Google struct {
	config      *oauth2.Config
	userInfoURL string
}

// GoogleUserInfo represents Google user info response.
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// NewGoogle creates a Google OAuth client.
func NewGoogle(clientID, clientSecret, redirectURL, tokenURL, userInfoURL string, scopes []string) *Google {
	return &Google{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint: oauth2.Endpoint{
				TokenURL: tokenURL,
			},
			Scopes: scopes,
		},
		userInfoURL: userInfoURL,
	}
}

// WithEndpoint sets a custom OAuth endpoint (for tests).
func (g *Google) WithEndpoint(endpoint oauth2.Endpoint) *Google {
	g.config.Endpoint = endpoint
	return g
}

// WithUserInfoURL sets a custom user info URL.
func (g *Google) WithUserInfoURL(userInfoURL string) *Google {
	g.userInfoURL = userInfoURL
	return g
}

// Exchange exchanges an auth code for a user profile.
func (g *Google) Exchange(ctx context.Context, code string) (*models.User, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google exchange failed: %w", err)
	}

	client := g.config.Client(ctx, token)

	resp, err := client.Get(g.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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
