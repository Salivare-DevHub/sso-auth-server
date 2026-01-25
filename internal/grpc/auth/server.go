package auth

import (
	"context"
	authv1 "github.com/Salivare-DevHub/protos-sso/gen/go/sso/auth"
	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/auth/providers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, provider string, code string, appID int64) (access, refresh string, err error)
}

type ServerAPI struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

func (s *ServerAPI) ExchangeCode(
	ctx context.Context,
	req *authv1.ExchangeCodeRequest,
) (*authv1.ExchangeCodeResponse, error) {

	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is empty")
	}

	provider := providerToString(req.GetProvider())
	if provider == "" {
		return nil, status.Error(codes.InvalidArgument, "unknown provider")
	}

	access, refresh, err := s.auth.Login(ctx, provider, req.GetCode(), req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to exchange code")
	}

	return &authv1.ExchangeCodeResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func providerToString(p authv1.Provider) string {
	switch p {
	case authv1.Provider_GOOGLE:
		return providers.ProviderGoogle
	case authv1.Provider_YANDEX:
		return providers.ProviderYandex
	default:
		return ""
	}
}
