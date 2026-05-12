package pkg

import (
	"context"

	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	testbinv1beta "github.com/datakit-dev/dtkt-integrations/testbin/pkg/proto/integration/testbin/v1beta"
)

type Instance struct {
	config *testbinv1beta.Config
}

func NewInstance(_ context.Context, config *testbinv1beta.Config) (*Instance, error) {
	return &Instance{config: config}, nil
}

func (i *Instance) CheckAuth(_ context.Context, _ *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "testbin does not require auth")
}

func (i *Instance) Close() error {
	return nil
}
