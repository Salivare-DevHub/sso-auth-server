package models

import "time"

// SessionData stores refresh session data.
type SessionData struct {
	UserID       string    `json:"user_id"`
	SessionID    string    `json:"session_id"`
	RefreshToken string    `json:"refresh_token"`
	AppID        string    `json:"app_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}
