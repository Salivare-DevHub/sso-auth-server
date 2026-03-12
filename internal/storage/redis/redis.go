package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

// New creates a Redis client.
func New(addr string) *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)
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
func (s *RefreshStore) Save(ctx context.Context, userID, sessionID, refresh, appID string) error {
	sessionKey := fmt.Sprintf("session:%s:%s", userID, sessionID)
	refreshKey := fmt.Sprintf("refresh:%s", refresh)

	data := models.SessionData{
		UserID:       userID,
		SessionID:    sessionID,
		RefreshToken: refresh,
		AppID:        appID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.ttl),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session data: %w", err)
	}

	pipe := s.client.TxPipeline()
	pipe.Set(ctx, sessionKey, jsonData, s.ttl)
	pipe.Set(ctx, refreshKey, sessionKey, s.ttl)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("save session data: %w", err)
	}

	return nil
}

// GetSession returns a session by sessionID.
func (s *RefreshStore) GetSession(ctx context.Context, userID string, sessionID string) (*models.SessionData, error) {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session models.SessionData
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

// GetByRefreshToken returns a session by refresh token.
func (s *RefreshStore) GetByRefreshToken(ctx context.Context, refresh string) (*models.SessionData, error) {
	refreshKey := fmt.Sprintf("refresh:%s", refresh)

	sessionKey, err := s.client.Get(ctx, refreshKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get refresh index: %w", err)
	}

	data, err := s.client.Get(ctx, sessionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session models.SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("unmarshal session data: %w", err)
	}

	return &session, nil
}

// DeleteByRefreshToken deletes a session by refresh token.
func (s *RefreshStore) DeleteByRefreshToken(ctx context.Context, refresh string) error {
	refreshKey := fmt.Sprintf("refresh:%s", refresh)

	sessionKey, err := s.client.Get(ctx, refreshKey).Result()
	if err != nil {
		return fmt.Errorf("get refresh index: %w", err)
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, refreshKey)
	pipe.Del(ctx, sessionKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete session data: %w", err)
	}

	return nil
}

// GetAllUserSessions returns all user sessions.
func (s *RefreshStore) GetAllUserSessions(ctx context.Context, userID string) ([]models.SessionData, error) {
	pattern := fmt.Sprintf("session:%s:*", userID)

	var sessions []models.SessionData
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

			var session models.SessionData
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
