package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

// ErrAppNotFound is returned when an app is not found.
// ErrInvalidToken is returned when a bearer token is invalid.
// ErrTokenNotFound is returned when a bearer token is missing.
var (
	ErrAppNotFound   = errors.New("application not found")
	ErrInvalidToken  = errors.New("invalid bearer token")
	ErrTokenNotFound = errors.New("bearer token not found")
)

// Manager provides app lookup and bearer token validation.
type Manager struct {
	apps map[string]*models.App
	mu   sync.RWMutex
}

// NewAppManager creates a new Manager.
func NewAppManager(appsMap map[string]*models.App) *Manager {
	return &Manager{
		apps: appsMap,
	}
}

// GetByID returns an app by ID.
func (m *Manager) GetByID(ctx context.Context, id string) (*models.App, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	app, ok := m.apps[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAppNotFound, id)
	}

	return app, nil
}

// VerifyBearerToken validates bearer token and returns app_id.
func (m *Manager) VerifyBearerToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrTokenNotFound
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for appID, app := range m.apps {
		if app.BearerToken == token {
			return appID, nil
		}
	}

	return "", ErrInvalidToken
}

// GetAppByBearerToken returns an app by bearer token.
func (m *Manager) GetAppByBearerToken(ctx context.Context, token string) (*models.App, error) {
	if token == "" {
		return nil, ErrTokenNotFound
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, app := range m.apps {
		if app.BearerToken == token {
			return app, nil
		}
	}

	return nil, ErrInvalidToken
}
