package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func NewAccessToken(userID string, app models.App) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = userID
	claims["app_id"] = app.ID
	claims["exp"] = time.Now().Add(app.AccessTTL).Unix()

	return token.SignedString([]byte(app.Secret))
}

func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
