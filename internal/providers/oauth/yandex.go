package oauth

import (
	"context"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

type Yandex struct {
	clientID     string
	clientSecret string
}

func NewYandex(clientID, clientSecret string) *Yandex {
	return &Yandex{clientID: clientID, clientSecret: clientSecret}
}

func (y *Yandex) Exchange(ctx context.Context, code string) (*models.User, error) {
	// вызов Yandex API
	panic("implement me")
}
