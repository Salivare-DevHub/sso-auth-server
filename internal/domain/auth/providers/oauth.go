package providers

import (
	"context"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
)

type OAuth interface {
	Exchange(ctx context.Context, code string) (*models.User, error)
}

type GoogleIdentity struct {
	oauth OAuth
}

func NewGoogleIdentity(o OAuth) *GoogleIdentity {
	return &GoogleIdentity{oauth: o}
}

func (g *GoogleIdentity) FetchUser(ctx context.Context, code string) (*models.User, error) {
	return g.oauth.Exchange(ctx, code)
}

type YandexIdentity struct {
	oauth OAuth
}

func NewYandexIdentity(o OAuth) *YandexIdentity {
	return &YandexIdentity{oauth: o}
}

func (y *YandexIdentity) FetchUser(ctx context.Context, code string) (*models.User, error) {
	return y.oauth.Exchange(ctx, code)
}
