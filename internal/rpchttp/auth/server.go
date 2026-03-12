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
	"github.com/salivare-io/sso-auth-server/internal/domain/models"
)

// Auth defines auth service operations.
type Auth interface {
	Login(ctx context.Context, provider string, code string, appID string) (access, refresh string, err error)
	UserDetails(ctx context.Context, accessToken string) (*models.User, error)
	Refresh(ctx context.Context, appID string, refreshToken string) (access, refresh string, err error)
	Logout(ctx context.Context, refreshToken string) error
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	GetPermissions(ctx context.Context, accessToken string) ([]models.Permission, error)
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

func (s *ServerAPI) UserDetails(
	ctx context.Context,
	req *connect.Request[authv1.UserDetailsRequest],
) (*connect.Response[authv1.UserDetailsResponse], error) {
	if req.Msg.AccessToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("access_token is required"))
	}

	user, err := s.auth.UserDetails(ctx, req.Msg.AccessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return &connect.Response[authv1.UserDetailsResponse]{
		Msg: &authv1.UserDetailsResponse{
			UserId: user.ID,
			Email:  user.Email,
			Name:   user.Name,
		},
	}, nil
}

func (s *ServerAPI) RefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RefreshTokenRequest],
) (*connect.Response[authv1.RefreshTokenResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh_token is required"))
	}

	appID := chimiddleware.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("app_id not found in context"))
	}

	access, refresh, err := s.auth.Refresh(ctx, appID, req.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return &connect.Response[authv1.RefreshTokenResponse]{
		Msg: &authv1.RefreshTokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
		},
	}, nil
}

func (s *ServerAPI) Logout(
	ctx context.Context,
	req *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh_token is required"))
	}

	if err := s.auth.Logout(ctx, req.Msg.RefreshToken); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return &connect.Response[authv1.LogoutResponse]{Msg: &authv1.LogoutResponse{}}, nil
}

func (s *ServerAPI) RevokeRefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RevokeRefreshTokenRequest],
) (*connect.Response[authv1.RevokeRefreshTokenResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh_token is required"))
	}

	if err := s.auth.RevokeRefreshToken(ctx, req.Msg.RefreshToken); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return &connect.Response[authv1.RevokeRefreshTokenResponse]{
		Msg: &authv1.RevokeRefreshTokenResponse{},
	}, nil
}

func (s *ServerAPI) GetPermissions(
	ctx context.Context,
	req *connect.Request[authv1.GetPermissionsRequest],
) (*connect.Response[authv1.GetPermissionsResponse], error) {
	if req.Msg.AccessToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("access_token is required"))
	}

	perms, err := s.auth.GetPermissions(ctx, req.Msg.AccessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	resp := &authv1.GetPermissionsResponse{}
	for _, perm := range perms {
		resp.Permissions = append(
			resp.Permissions, &authv1.Permission{
				Code:        perm.Code,
				Description: perm.Description,
			},
		)
	}

	return &connect.Response[authv1.GetPermissionsResponse]{Msg: resp}, nil
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
