package grpc

import (
	"context"
	"errors"
	"google.golang.org/grpc"
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
