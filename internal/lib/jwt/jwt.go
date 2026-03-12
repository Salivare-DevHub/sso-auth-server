package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

// Claims defines custom JWT claims.
type Claims struct {
	jwt.RegisteredClaims
	UID       string `json:"uid"`
	SessionID string `json:"session_id"`
	AppID     string `json:"app_id"`
}

// NewAccessToken creates and signs a new access token.
func NewAccessToken(userID string, sessionID string, app models.App) (string, error) {
	tokenID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}
	if app.SigningKey == "" {
		return "", fmt.Errorf("missing signing key for app %s", app.ID)
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(app.AccessTTL)),
		},
		UID:       userID,
		SessionID: sessionID,
		AppID:     app.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(app.SigningKey))
}

// NewRefreshToken generates a secure refresh token.
func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
