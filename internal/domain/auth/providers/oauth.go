package providers

import (
	"context"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

// ProviderGoogle and ProviderYandex are supported provider names.
const (
	ProviderGoogle = "google"
	ProviderYandex = "yandex"
)

// OAuth defines an OAuth exchange interface.
type OAuth interface {
	Exchange(ctx context.Context, code string) (*models.User, error)
}

// GoogleIdentity fetches users from Google OAuth.
type GoogleIdentity struct {
	oauth OAuth
}

// NewGoogleIdentity creates a GoogleIdentity.
func NewGoogleIdentity(o OAuth) *GoogleIdentity {
	return &GoogleIdentity{oauth: o}
}

// FetchUser returns a user from Google OAuth.
func (g *GoogleIdentity) FetchUser(ctx context.Context, code string) (*models.User, error) {
	return g.oauth.Exchange(ctx, code)
}

// YandexIdentity fetches users from Yandex OAuth.
type YandexIdentity struct {
	oauth OAuth
}

// NewYandexIdentity creates a YandexIdentity.
func NewYandexIdentity(o OAuth) *YandexIdentity {
	return &YandexIdentity{oauth: o}
}

// FetchUser returns a user from Yandex OAuth.
func (y *YandexIdentity) FetchUser(ctx context.Context, code string) (*models.User, error) {
	return y.oauth.Exchange(ctx, code)
}
