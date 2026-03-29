package pkg

import (
	"context"
	"fmt"

	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/middleware"
	fb "github.com/huandu/facebook/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var DefaultScopes = []string{
	"business_management",
	"ads_management",
	"ads_read",
	"pages_manage_posts",
	"pages_show_list",
	"pages_read_engagement",
	"pages_read_user_content",
	"read_insights",
	"instagram_basic",
}

type (
	Instance struct {
		config *Config
		app    *fb.App
		client *Client
	}
	Config struct {
		AppID      string `json:"app_id"`
		AppSecret  string `json:"app_secret"`
		APIVersion string `json:"api_version"`
		ConfigID   string `json:"config_id,omitempty"`
	}
)

// NewInstance creates a new service instance
func NewInstance(ctx context.Context, config *Config) (*Instance, error) {
	logger := log.FromCtx(ctx)

	var (
		app    = NewApp(config)
		client = NewClient(logger, app, config)
	)

	// sourceService, err := fivetran.NewSourceService(ctx,
	// 	NewFacebookAdsSource(config, app),
	// )
	// if err != nil {
	// 	return nil, err
	// }

	return &Instance{
		// SourceService: sourceService,
		config: config,
		app:    app,
		client: client,
	}, nil
}

// Close is called to release resources used by the service (e.g. client connections)
func (s *Instance) Close() error {
	return status.Errorf(codes.Unimplemented, "method Close not implemented")
}

func (s *Instance) CheckAuth(ctx context.Context, req *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	resp := &basev1beta1.CheckAuthResponse{
		Type: sharedv1beta1.AuthType_AUTH_TYPE_OAUTH,
	}

	switch req.Type {
	case sharedv1beta1.AuthCheck_AUTH_CHECK_OAUTH_CODE:
		config, opts := middleware.OAuthConfigFromProto(NewAuthConfig(s.config, DefaultScopes))
		config.RedirectURL = req.GetCodeRequest().RedirectUrl

		resp.Oauth = &basev1beta1.CheckAuthResponse_AuthCodeUrl{
			AuthCodeUrl: config.AuthCodeURL(req.GetCodeRequest().State, opts...),
		}
	case sharedv1beta1.AuthCheck_AUTH_CHECK_OAUTH_CALLBACK:
		if req.GetTokenRequest() == nil {
			resp.Success = false
			resp.Error = fmt.Sprintf("check auth %s: missing token request", sharedv1beta1.AuthCheck_AUTH_CHECK_OAUTH_CALLBACK)
			return resp, nil
		}

		config, opts := middleware.OAuthConfigFromProto(NewAuthConfig(s.config, DefaultScopes))
		config.RedirectURL = req.GetTokenRequest().RedirectUrl

		token, err := config.Exchange(ctx, req.GetTokenRequest().Code, opts...)
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
			return resp, nil
		}

		resp.Oauth = &basev1beta1.CheckAuthResponse_Token{
			Token: middleware.OAuthTokenToProto(token),
		}
	case sharedv1beta1.AuthCheck_AUTH_CHECK_OAUTH_REFRESH:
		if req.GetRefreshRequest() == nil {
			resp.Success = false
			resp.Error = fmt.Sprintf("check auth %s: missing token request", sharedv1beta1.AuthCheck_AUTH_CHECK_OAUTH_REFRESH)
			return resp, nil
		}

		// token, err := RefreshToken(ctx, s.app, req.GetRefreshRequest().RefreshToken)
		// if err != nil {
		// 	resp.Success = false
		// 	resp.Error = err.Error()
		// 	return resp, nil
		// }

		// resp.Oauth = &basev1beta1.CheckAuthResponse_Token{
		// 	Token: middleware.OAuthTokenToProto(token),
		// }
	}

	resp.Success = true
	return resp, nil
}
