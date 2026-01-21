package oauth

import (
	"context"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
)

type YandexOAuth struct {
	clientID     string
	clientSecret string
}

func NewYandexOAuth(clientID, clientSecret string) *YandexOAuth {
	return &YandexOAuth{clientID: clientID, clientSecret: clientSecret}
}

func (y *YandexOAuth) Exchange(ctx context.Context, code string) (*models.User, error) {
	// вызов Yandex API
	panic("implement me")
}
