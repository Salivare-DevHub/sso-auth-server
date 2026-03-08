package user

import (
	"context"
	"errors"
	"sync"
)

type Client struct {
	mu    sync.Mutex
	seq   int64
	users map[string]int64
}

func NewClient() *Client {
	return &Client{
		users: make(map[string]int64),
		seq:   0,
	}
}

func (s *Client) SaveUser(ctx context.Context, email string) (int64, error) {
	if email == "" {
		return 0, errors.New("email is empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.users[email]; ok {
		return id, nil
	}

	s.seq++
	s.users[email] = s.seq

	return s.seq, nil
}
