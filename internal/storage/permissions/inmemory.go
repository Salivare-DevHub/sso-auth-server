// Package permissions provides permission storage implementations.
package permissions

import (
	"context"
	"sync"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

// InMemoryStore provides a stub permissions store.
type InMemoryStore struct {
	mu    sync.RWMutex
	perms map[string][]models.Permission
}

// NewInMemoryStore creates an empty in-memory permissions store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		perms: make(map[string][]models.Permission),
	}
}

// GetByUserID returns permissions for a user.
func (s *InMemoryStore) GetByUserID(ctx context.Context, userID string) ([]models.Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	perms := s.perms[userID]
	if perms == nil {
		return []models.Permission{}, nil
	}

	out := make([]models.Permission, len(perms))
	copy(out, perms)
	return out, nil
}
