package auth

import (
	"context"
	ssov1 "github.com/Salivare-DevHub/protos-sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	StartAuth(ctx context.Context, provider string) (redirectURL string, err error)
	ExchangeCode(ctx context.Context, provider string, code string) (access string, refresh string, err error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

// StartAuth returns the redirect URL to start OAuth2 authorization.
func (s *ServerAPI) StartAuth(
	ctx context.Context,
	req *ssov1.StartAuthRequest,
) (*ssov1.StartAuthResponse, error) {

	if req.GetProvider() == "" {
		return nil, status.Error(codes.InvalidArgument, "provider is empty")
	}

	url, err := s.auth.StartAuth(ctx, req.GetProvider())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to start auth")
	}

	return &ssov1.StartAuthResponse{
		RedirectUrl: url,
	}, nil
}

// ExchangeCode exchange OAuth2 code for user profile
func (s *ServerAPI) ExchangeCode(
	ctx context.Context,
	req *ssov1.ExchangeCodeRequest,
) (*ssov1.ExchangeCodeResponse, error) {

	if req.GetProvider() == "" {
		return nil, status.Error(codes.InvalidArgument, "provider is empty")
	}

	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is empty")
	}

	access, refresh, err := s.auth.ExchangeCode(ctx, req.GetProvider(), req.GetCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to exchange code")
	}

	return &ssov1.ExchangeCodeResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
