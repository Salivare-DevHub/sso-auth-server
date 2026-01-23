package oauth

import (
	"context"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
)

type Google struct {
	clientID     string
	clientSecret string
}

func NewGoogle(clientID, clientSecret string) *Google {
	return &Google{clientID: clientID, clientSecret: clientSecret}
}

func (g *Google) Exchange(ctx context.Context, code string) (*models.User, error) {
	// вызов Google API
	panic("implement me")
}
