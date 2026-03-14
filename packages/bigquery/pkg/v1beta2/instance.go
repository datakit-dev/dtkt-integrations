package v1beta2

import (
	"context"
	"sync"

	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	bigqueryv1beta "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/proto/integration/bigquery/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

var Package *sharedv1beta1.Package

type Instance struct {
	basev1beta1.UnimplementedBaseServiceServer

	config *bigqueryv1beta.Config
	client *lib.Client
	mut    sync.Mutex
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
		config: config,
		client: client,
	}, nil
}

func (i *Instance) Client() *lib.Client {
	return i.client
}

func (i *Instance) GetAuditLogEvent(ctx context.Context, auditEvent v1beta1.RegisteredEvent, auditLog *audit.AuditLog) (any, error) {
	return NewAuditLogEvent(ctx, auditEvent, auditLog)
}

// func (i *Instance) GetFivetranDestination(ctx context.Context) (*fivetran.Destination, error) {
// 	i.mut.Lock()
// 	defer i.mut.Unlock()

// 	if i.getDestination == nil {
// 		i.getDestination = sync.OnceValues(func() (*fivetran.Destination, error) {
// 			if i.config.FivetranConfig == nil {
// 				return nil, fmt.Errorf("fivetran replication not supported: config missing %q", "fivetran_replication")
// 			} else if i.client.CredsJson == nil {
// 				return nil, fmt.Errorf("fivetran replication not supported: config missing %q", "credentials_json")
// 			}

// 			client := &fivetranv1.ClientConfig{}
// 			err := i.config.FivetranConfig.Client.UnmarshalTo(client)
// 			if err != nil {
// 				return nil, err
// 			}

// 			config := &fivetranv1.DestinationConfig{}
// 			err = i.config.FivetranConfig.Destination.UnmarshalTo(config)
// 			if err != nil {
// 				return nil, err
// 			}

// 			config.Attributes = &fivetranv1.DestinationConfig_Attributes{
// 				ProjectId:       new(i.config.ProjectId),
// 				DataSetLocation: new(i.config.Location),
// 				Bucket:          new(i.config.GcsBucket),
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

// 	return i.getDestination()
// }

func (i *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return &basev1beta1.CheckAuthResponse{}, nil
}

func (i *Instance) Close() error {
	return i.client.Close()
}
