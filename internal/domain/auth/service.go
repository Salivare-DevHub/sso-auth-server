package auth

import (
	"context"
	"errors"
	"log/slog"
)

type LoginStrategy interface {
	Login(ctx context.Context, code string) (access, refresh string, err error)
}

var (
	ErrUnknownProvider = errors.New("unknown provider")
)

type Service struct {
	log        *slog.Logger
	strategies map[string]LoginStrategy
}

func New(log *slog.Logger, strategies map[string]LoginStrategy) *Service {
	return &Service{
		log:        log,
		strategies: strategies,
	}
}

func (s *Service) Login(ctx context.Context, provider string, code string) (string, string, error) {
	strat, ok := s.strategies[provider]
	if !ok {
		return "", "", ErrUnknownProvider
	}

	return strat.Login(ctx, code)
}
