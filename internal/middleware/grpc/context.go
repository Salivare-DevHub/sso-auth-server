package grpc

import (
	"context"
	"errors"
)

type ctxKey string

const (
	AppIDKey   ctxKey = "app_id"
	AppNameKey ctxKey = "app_name"
)

var (
	ErrMissingAppID   = errors.New("missing app_id in context")
	ErrInvalidAppID   = errors.New("invalid app_id type")
	ErrMissingAppName = errors.New("missing app_name in context")
	ErrInvalidAppName = errors.New("invalid app_name type")
)

func GetAppID(ctx context.Context) (string, error) {
	v := ctx.Value(AppIDKey)
	if v == nil {
		return "", ErrMissingAppID
	}

	id, ok := v.(string)
	if !ok {
		return "", ErrInvalidAppID
	}

	return id, nil
}

func GetAppName(ctx context.Context) (string, error) {
	v := ctx.Value(AppNameKey)
	if v == nil {
		return "", ErrMissingAppName
	}

	name, ok := v.(string)
	if !ok {
		return "", ErrInvalidAppName
	}

	return name, nil
}
