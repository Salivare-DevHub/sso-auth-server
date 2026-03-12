package user

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// User represents a stored user.
type User struct {
	ID    string
	Email string
	Name  string
}

// InMemoryRepository stores users in memory.
type InMemoryRepository struct {
	mu    sync.Mutex
	users map[string]*User  // userID -> User
	email map[string]string // email -> userID
}

// NewInMemoryRepository creates an in-memory repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		users: make(map[string]*User),
		email: make(map[string]string),
	}
}

// SaveUser creates a new user or returns an existing one.
func (r *InMemoryRepository) SaveUser(ctx context.Context, email string) (string, error) {
	if email == "" {
		return "", errors.New("email is empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// If the user already exists, return its ID.
	if userID, ok := r.email[email]; ok {
		return userID, nil
	}

	// Generate a new UUID7 for the user.
	userID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	userIDStr := userID.String()
	user := &User{
		ID:    userIDStr,
		Email: email,
	}

	r.users[userIDStr] = user
	r.email[email] = userIDStr

	return userIDStr, nil
}

// GetByID returns a user by ID.
func (r *InMemoryRepository) GetByID(ctx context.Context, userID string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetByEmail returns a user by email.
func (r *InMemoryRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userID, ok := r.email[email]
	if !ok {
		return nil, ErrUserNotFound
	}

	user := r.users[userID]
	return user, nil
}
