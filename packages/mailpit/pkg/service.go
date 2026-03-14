package pkg

import (
	"context"
	"fmt"

	oapigen "github.com/datakit-dev/dtkt-integrations/mailpit/pkg/oapi/gen"

	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
)

type Config struct {
	Templates []*emailv1beta1.EmailTemplate `json:"templates,omitempty"`
}

type Instance struct {
	config *Config
	client *oapigen.Client

	template *TemplateService
}

// NewInstance creates a new service instance
func NewInstance(ctx context.Context, config *Config) (*Instance, error) {
	client, err := oapigen.NewClient("http://localhost:8025")
	if err != nil {
		return nil, fmt.Errorf("failed to create Mailpit client: %v", err)
	}

	template, err := NewTemplateService(config.Templates)
	if err != nil {
		return nil, err
	}

	return &Instance{
		config: config,
		client: client,

		template: template,
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
