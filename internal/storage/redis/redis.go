package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// New creates a Redis client.
func New(addr string) *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)
}

// SessionData stores refresh session data.
type SessionData struct {
	UserID       string    `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshStore manages refresh tokens in Redis.
type RefreshStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRefreshStore creates a RefreshStore with a TTL.
func NewRefreshStore(client *redis.Client, ttl time.Duration) *RefreshStore {
	return &RefreshStore{
		client: client,
		ttl:    ttl,
	}
}

// Save stores refresh token data bound to sessionID.
// Key: session:<userID>:<sessionID>
// Stores full session info.
func (s *RefreshStore) Save(ctx context.Context, userID string, sessionID string, refresh string) error {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)

	data := SessionData{
		UserID:       userID,
		RefreshToken: refresh,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.ttl),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session data: %w", err)
	}

	return s.client.Set(ctx, key, jsonData, s.ttl).Err()
}

// GetSession returns a session by sessionID.
func (s *RefreshStore) GetSession(ctx context.Context, userID string, sessionID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("unmarshal session data: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session.
func (s *RefreshStore) DeleteSession(ctx context.Context, userID string, sessionID string) error {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	return s.client.Del(ctx, key).Err()
}

// GetAllUserSessions returns all user sessions.
func (s *RefreshStore) GetAllUserSessions(ctx context.Context, userID string) ([]SessionData, error) {
	pattern := fmt.Sprintf("session:%s:*", userID)

	var sessions []SessionData
	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("scan keys: %w", err)
		}

		for _, key := range keys {
			data, err := s.client.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var session SessionData
			if err := json.Unmarshal([]byte(data), &session); err != nil {
				continue
			}

			sessions = append(sessions, session)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return sessions, nil
}
