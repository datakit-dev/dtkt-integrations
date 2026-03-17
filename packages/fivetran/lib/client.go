package fivetran

import (
	"context"
	"fmt"

	fivetranv1 "github.com/datakit-dev/dtkt-integrations/fivetran/gen/integration/fivetran/v1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/fivetran/go-fivetran"
)

type (
	Client struct {
		*fivetran.Client
		config *fivetranv1.Config
	}
	InstanceWithCredentials interface {
		v1beta1.InstanceType
		GetFivetranConfig(context.Context) (*fivetranv1.Config, error)
	}
)

func NewClient(config *fivetranv1.Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if config.ApiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	} else if config.ApiSecret == "" {
		return nil, fmt.Errorf("api_secret is required")
	}

	return &Client{
		Client: fivetran.New(config.ApiKey, config.ApiSecret),
		config: config,
	}, nil
}

func GetClientFromInstance[I InstanceWithCredentials](ctx context.Context, mux v1beta1.InstanceMux[I]) (*Client, error) {
	inst, err := mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	config, err := inst.GetFivetranConfig(ctx)
	if err != nil {
		return nil, err
	}

	return NewClient(config)
}
