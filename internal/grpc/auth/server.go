package auth

import (
	"context"
	ssov1 "github.com/Salivare-DevHub/protos-sso/gen/go/sso"
	"google.golang.org/grpc"
)

type Auth interface {
	StartAuth(ctx context.Context, req *ssov1.StartAuthRequest) (*ssov1.StartAuthResponse, error)
	ExchangeCode(ctx context.Context, req *ssov1.ExchangeCodeRequest) (*ssov1.ExchangeCodeResponse, error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

// StartAuth returns the redirect URL to start OAuth2 authorization.
func (s *ServerAPI) StartAuth(ctx context.Context, req *ssov1.StartAuthRequest) (*ssov1.StartAuthResponse, error) {
	return s.auth.StartAuth(ctx, req)
}

// ExchangeCode exchange OAuth2 code for user profile
func (s *ServerAPI) ExchangeCode(ctx context.Context, req *ssov1.ExchangeCodeRequest) (*ssov1.ExchangeCodeResponse, error) {
	return s.auth.ExchangeCode(ctx, req)
}
