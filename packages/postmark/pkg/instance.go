package pkg

import (
	"context"
	"fmt"

	accountoapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/accountoapi/gen"
	serveroapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi/gen"

	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
)

type Config struct {
	AccountApiToken string `json:"account_api_token"`
	ServerApiToken  string `json:"server_api_token"`
}

type Instance struct {
	config        *Config
	accountClient *accountoapigen.Client
	serverClient  *serveroapigen.Client
}

// NewInstance creates a new service instance
func NewInstance(ctx context.Context, config *Config) (*Instance, error) {
	var accountClient *accountoapigen.Client
	if config.AccountApiToken != "" {
		ac, err := accountoapigen.NewClient("https://api.postmarkapp.com")

		if err != nil {
			return nil, fmt.Errorf("failed to create Postmark Account client: %v", err)
		}

		accountClient = ac
	}

	serverClient, err := serveroapigen.NewClient("https://api.postmarkapp.com")
	if err != nil {
		return nil, fmt.Errorf("failed to create Postmark Server client: %v", err)
	}

	return &Instance{
		config: config,

		accountClient: accountClient,
		serverClient:  serverClient,
	}, nil
}

func (s *Instance) Close() error {
	return nil
}

func (s *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return &basev1beta1.CheckAuthResponse{
		Type: sharedv1beta1.AuthType_AUTH_TYPE_UNSPECIFIED,
	}, nil
}
