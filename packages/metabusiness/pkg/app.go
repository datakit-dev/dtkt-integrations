package pkg

import (
	"context"
	"fmt"
	"time"

	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"

	fb "github.com/huandu/facebook/v2"
	"golang.org/x/oauth2"
)

func NewApp(config *Config) *fb.App {
	var app = fb.New(config.AppID, config.AppSecret)
	app.EnableAppsecretProof = true
	fb.Version = config.APIVersion
	return app
}

func NewAuthConfig(config *Config, scopes []string) *sharedv1beta1.OAuthConfig {
	var params = map[string]string{
		"access_type": "offline",
	}

	if config.ConfigID != "" {
		params["config_id"] = config.ConfigID
		params["override_default_response_type"] = "true"
		scopes = []string{}
	}

	return &sharedv1beta1.OAuthConfig{
		ClientId:     config.AppID,
		ClientSecret: config.AppSecret,
		Endpoint: &sharedv1beta1.OAuthEndpoint{
			AuthUrl:  fmt.Sprintf("https://www.facebook.com/%s/dialog/oauth", config.APIVersion),
			TokenUrl: fmt.Sprintf("https://graph.facebook.com/%s/oauth/access_token", config.APIVersion),
		},
		Params: params,
		Scopes: scopes,
	}
}

func NewSession(ctx context.Context, app *fb.App, oauthConfig *sharedv1beta1.OAuthConfig) (*fb.Session, error) {
	// token, err := middleware.OAuthTokenFromContext(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// client, err := middleware.OAuthClientFromContext(ctx, oauthConfig)
	// if err != nil {
	// 	return nil, err
	// }

	// session := app.Session(token.AccessToken).WithContext(ctx)
	// session.HttpClient = client

	// if err := session.EnableAppsecretProof(true); err != nil {
	// 	return nil, err
	// }

	// return session, nil
	return nil, nil
}

func RefreshToken(ctx context.Context, app *fb.App, token *sharedv1beta1.OAuthToken) (*oauth2.Token, error) {
	if token == nil {
		return nil, fmt.Errorf("rotate token missing")
	}

	accessToken, expiresIn, err := app.ExchangeToken(token.AccessToken)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: accessToken,
		ExpiresIn:   int64(expiresIn),
		Expiry:      time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}
