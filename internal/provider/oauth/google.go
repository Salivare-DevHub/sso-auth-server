package oauth

import (
	"context"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
)

type GoogleOAuth struct {
	clientID     string
	clientSecret string
}

func NewGoogleOAuth(clientID, clientSecret string) *GoogleOAuth {
	return &GoogleOAuth{clientID: clientID, clientSecret: clientSecret}
}

func (g *GoogleOAuth) Exchange(ctx context.Context, code string) (*models.User, error) {
	// вызов Google API
	panic("implement me")
}
