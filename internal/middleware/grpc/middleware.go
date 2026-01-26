package grpc

import (
	"context"
	"errors"
	"github.com/Salivare-DevHub/sso-auth-server/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

func Timeout(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := handler(ctx, req)

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, status.Error(408, "request timeout")
		}

		return resp, err
	}
}

func InternalAuth(apps map[string]config.AppCredentials) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		clientIDs := md.Get("x-client-id")
		secrets := md.Get("x-client-secret")

		if len(clientIDs) != 1 || len(secrets) != 1 {
			return nil, status.Error(codes.Unauthenticated, "missing credentials")
		}

		clientID := clientIDs[0]
		secret := secrets[0]

		if clientID == "" || secret == "" {
			return nil, status.Error(codes.Unauthenticated, "empty credentials")
		}

		for appName, creds := range apps {
			if creds.ClientID == clientID && creds.ClientSecret == secret {
				ctx = context.WithValue(ctx, AppIDKey, clientID)
				ctx = context.WithValue(ctx, AppNameKey, appName)

				return handler(ctx, req)
			}
		}

		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
}
