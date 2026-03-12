package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
	"golang.org/x/oauth2"
)

// Yandex implements OAuth for Yandex.
type Yandex struct {
	config      *oauth2.Config
	userInfoURL string
}

// YandexUserInfo represents Yandex user info response.
type YandexUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RealName  string `json:"real_name"`
	Email     string `json:"default_email"`
	Picture   string `json:"default_avatar_id"`
}

// NewYandex creates a Yandex OAuth client.
func NewYandex(clientID, clientSecret, redirectURL, tokenURL, userInfoURL string, scopes []string) *Yandex {
	return &Yandex{
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
func (y *Yandex) WithEndpoint(endpoint oauth2.Endpoint) *Yandex {
	y.config.Endpoint = endpoint
	return y
}

// WithUserInfoURL sets a custom user info URL.
func (y *Yandex) WithUserInfoURL(userInfoURL string) *Yandex {
	y.userInfoURL = userInfoURL
	return y
}

// Exchange exchanges an auth code for a user profile.
func (y *Yandex) Exchange(ctx context.Context, code string) (*models.User, error) {
	token, err := y.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("yandex exchange failed: %w", err)
	}

	client := y.config.Client(ctx, token)

	resp, err := client.Get(y.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex api returned status: %d", resp.StatusCode)
	}

	var userInfo YandexUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Build display name.
	name := userInfo.RealName
	if name == "" {
		name = userInfo.FirstName
		if userInfo.LastName != "" {
			name = name + " " + userInfo.LastName
		}
	}

	return &models.User{
		Email: userInfo.Email,
		Name:  name,
		ID:    fmt.Sprintf("%d", userInfo.ID), // Yandex uses a numeric ID.
	}, nil
}
