package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"

	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type (
	Instance struct {
		config   *Config
		client   openai.Client
		handlers util.SyncMap[string, *RealtimeHandler]
	}
	Config struct {
		APIKey    string `json:"api_key"`
		OrgID     string `json:"org_id,omitempty"`
		ProjectID string `json:"project_id,omitempty"`
	}
)

func NewInstance(ctx context.Context, config *Config) (*Instance, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(config.APIKey),
	}

	if config.OrgID != "" {
		opts = append(opts, option.WithOrganization(config.OrgID))
	}

	if config.ProjectID != "" {
		opts = append(opts, option.WithProject(config.ProjectID))
	}

	client := openai.NewClient(opts...)
	_, err := client.Models.List(ctx)
	if err != nil {
		return nil, err
	}

	return &Instance{
		config: config,
		client: client,
	}, nil
}

func (s *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return &basev1beta1.CheckAuthResponse{}, nil
}

// Close closes the service
func (s *Instance) Close() error {
	return nil
}
