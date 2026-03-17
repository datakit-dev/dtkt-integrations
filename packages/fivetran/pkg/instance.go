package pkg

import (
	"context"

	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	fivetranv1 "github.com/datakit-dev/dtkt-integrations/fivetran/gen/integration/fivetran/v1"
)

// Integration instance struct
type Instance struct {
	config *fivetranv1.Config
}

// Creates a new instance
func NewInstance(ctx context.Context, config *fivetranv1.Config) (*Instance, error) {
	return &Instance{
		config: config,
	}, nil
}

// Orchestrates OAuth checks/exchanges.
func (i *Instance) CheckAuth(ctx context.Context, req *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

// Close is called to release instance resources (e.g. client connections)
func (i *Instance) Close() error {
	return nil
}
