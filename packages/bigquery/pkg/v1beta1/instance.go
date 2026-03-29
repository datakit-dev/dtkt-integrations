package v1beta1

import (
	"context"
	"fmt"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"

	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	bigqueryv1beta "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/proto/integration/bigquery/v1beta"
	fivetranv1 "github.com/datakit-dev/dtkt-integrations/fivetran/gen/integration/fivetran/v1"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Instance struct {
	client *lib.Client
}

func NewInstance(ctx context.Context, config *bigqueryv1beta.Config) (*Instance, error) {
	return NewInstanceWithOptions(ctx, config)
}

func NewInstanceWithOptions(ctx context.Context, config *bigqueryv1beta.Config, opts ...option.ClientOption) (*Instance, error) {
	client, err := lib.NewClient(ctx, config, opts...)
	if err != nil {
		return nil, err
	}

	return &Instance{
		client: client,
	}, nil
}

func (i *Instance) Client() *lib.Client {
	return i.client
}

func (i *Instance) GetAuditLogEvent(ctx context.Context, event v1beta1.RegisteredEvent, log *audit.AuditLog) (any, error) {
	return NewAuditLogEvent(ctx, event, log)
}

func (i *Instance) GetFivetranConfig(context.Context) (*fivetranv1.Config, error) {
	if i.client.Config.GetFivetranReplication() == nil {
		return nil, fmt.Errorf("fivetran replication not configured")
	}

	return i.client.Config.GetFivetranReplication().GetConfig(), nil
}

func (i *Instance) GetFivetranDestinations(ctx context.Context) ([]*fivetranv1.Destination, error) {
	if i.client.Config.GetFivetranReplication() == nil {
		return nil, fmt.Errorf("fivetran replication not configured")
	}

	return []*fivetranv1.Destination{
		i.client.Config.GetFivetranReplication().GetDestination(),
	}, nil
}

func (i *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

func (i *Instance) Close() error {
	return i.client.Close()
}
