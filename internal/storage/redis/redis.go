package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func New(addr string) *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)
}

type RefreshStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRefreshStore(client *redis.Client, ttl time.Duration) *RefreshStore {
	return &RefreshStore{
		client: client,
		ttl:    ttl,
	}
}

func (s *RefreshStore) Save(ctx context.Context, userID int64, refresh string) error {
	key := fmt.Sprintf("refresh:%d", userID)

	return s.client.Set(ctx, key, refresh, s.ttl).Err()
}
