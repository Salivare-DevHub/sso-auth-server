package auth

import (
	"context"
	"errors"
	"net/http"

	connect "connectrpc.com/connect"

	chimiddleware "github.com/salivare-io/middleware/chi"
	authv1 "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1"
	authconnect "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1/authservicev1connect"

	"github.com/salivare-io/sso-auth-server/internal/domain/auth/providers"
)

// Auth defines auth service operations.
type Auth interface {
	Login(ctx context.Context, provider string, code string, appID string) (access, refresh string, err error)
}

// ServerAPI implements the generated Connect handler interface.
type ServerAPI struct {
	auth Auth
}

// compile-time assertion: ensure ServerAPI implements the generated Connect handler interface.
var _ authconnect.AuthServiceHandler = (*ServerAPI)(nil)

// NewServerAPI creates a new ServerAPI.
func NewServerAPI(authSvc Auth) *ServerAPI {
	return &ServerAPI{auth: authSvc}
}

// Register mounts the Connect handler in a router.
func Register(r interface{ Mount(string, http.Handler) }, authSvc Auth) {
	path, handler := authconnect.NewAuthServiceHandler(NewServerAPI(authSvc))
	r.Mount(path, handler)
}

func (s *ServerAPI) ExchangeCode(
	ctx context.Context,
	req *connect.Request[authv1.ExchangeCodeRequest],
) (*connect.Response[authv1.ExchangeCodeResponse], error) {
	// Get app_id from the context.
	appID := chimiddleware.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("app_id not found in context"))
	}

	providerStr := providerToString(req.Msg.Provider)
	if providerStr == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid provider"))
	}

	access, refresh, err := s.auth.Login(ctx, providerStr, req.Msg.Code, appID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return &connect.Response[authv1.ExchangeCodeResponse]{
		Msg: &authv1.ExchangeCodeResponse{
			AccessToken:  access,
			RefreshToken: refresh,
		},
	}, nil
}

// ...existing code...
func (s *ServerAPI) UserDetails(
	ctx context.Context,
	req *connect.Request[authv1.UserDetailsRequest],
) (*connect.Response[authv1.UserDetailsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("UserDetails: not implemented"))
}

func (s *ServerAPI) RefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RefreshTokenRequest],
) (*connect.Response[authv1.RefreshTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("RefreshToken: not implemented"))
}

func (s *ServerAPI) Logout(
	ctx context.Context,
	req *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("Logout: not implemented"))
}

func (s *ServerAPI) RevokeRefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RevokeRefreshTokenRequest],
) (*connect.Response[authv1.RevokeRefreshTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("RevokeRefreshToken: not implemented"))
}

func (s *ServerAPI) GetPermissions(
	ctx context.Context,
	req *connect.Request[authv1.GetPermissionsRequest],
) (*connect.Response[authv1.GetPermissionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("GetPermissions: not implemented"))
}

func providerToString(p authv1.Provider) string {
	switch p {
	case authv1.Provider_PROVIDER_GOOGLE:
		return providers.ProviderGoogle
	case authv1.Provider_PROVIDER_YANDEX:
		return providers.ProviderYandex
	default:
		return ""
	}
}
