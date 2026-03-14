package v1beta1

import (
	"context"
	"fmt"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	fivetranv1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/lib/fivetran/v1"

	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	bigqueryv1beta "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/proto/integration/bigquery/v1beta"
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

func (i *Instance) GetFivetranCredentials(context.Context) (creds *fivetranv1.Credentials, err error) {
	if i.client.Config.GetFivetran() == nil || i.client.Config.GetFivetran().GetCredentials() == nil {
		return nil, fmt.Errorf("fivetran credentials not found")
	}

	creds = new(fivetranv1.Credentials)
	err = i.client.Config.GetFivetran().GetCredentials().UnmarshalTo(creds)

	return
}

func (i *Instance) GetFivetranDestination(ctx context.Context, typeId string) (*fivetranv1.Destination, error) {
	if i.client.Config.GetFivetran() == nil || i.client.Config.GetFivetran().GetDestination() == nil {
		return nil, fmt.Errorf("fivetran destination not found")
	}

	dest := new(fivetranv1.Destination)
	err := i.client.Config.GetFivetran().GetDestination().UnmarshalTo(dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

// func (i *Instance) GetFivetranDestination(ctx context.Context) (*fivetran.Destination, error) {
// 	i.mut.Lock()
// 	defer i.mut.Unlock()

// 	if i.fivetranDest == nil {
// 		i.fivetranDest = sync.OnceValues(func() (*fivetran.Destination, error) {
// 			if i.client.Config.FivetranConfig == nil {
// 				return nil, fmt.Errorf("fivetran replication not supported: missing config")
// 			} else if i.client.CredsJson == nil {
// 				return nil, fmt.Errorf("fivetran replication not supported: config uses application default credentials, must provide credentials_json")
// 			}

// 			client := &fivetranv1.ClientConfig{}
// 			err := i.client.Config.FivetranConfig.Client.UnmarshalTo(client)
// 			if err != nil {
// 				return nil, err
// 			}

// 			config := &fivetranv1.DestinationConfig{}
// 			err = i.client.Config.FivetranConfig.Destination.UnmarshalTo(config)
// 			if err != nil {
// 				return nil, err
// 			}

// 			config.Attributes = &fivetranv1.DestinationConfig_Attributes{
// 				ProjectId:       new(i.client.Config.ProjectId),
// 				DataSetLocation: new(i.client.Config.Location),
// 				Bucket:          new(i.client.Config.GcsBucket),
// 				SecretKey:       new(string(i.client.CredsJson)),
// 			}

// 			return &fivetran.Destination{
// 				Type: &replicationv1beta1.DestinationType{
// 					Id:          "big_query",
// 					Name:        Package.GetIdentity().GetName(),
// 					Description: Package.GetDescription(),
// 					IconUrl:     Package.GetIcon(),
// 					Category:    "Warehouse",
// 				},
// 				Client: client,
// 				Config: config,
// 			}, nil
// 		})
// 	}

// 	return i.fivetranDest()
// }

func (i *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

func (i *Instance) Close() error {
	return i.client.Close()
}
