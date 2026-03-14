package v1beta2

import (
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta2 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta2"
)

type SchemaService struct {
	catalogv1beta2.UnimplementedSchemaServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewSchemaService(mux v1beta1.InstanceMux[*Instance]) *SchemaService {
	return &SchemaService{
		mux: mux,
	}
}

// func (s *SchemaService) CreateSchema(ctx context.Context, req *catalogv1beta2.CreateSchemaRequest) (*catalogv1beta2.CreateSchemaResponse, error) {
// 	return nil, nil
// }

// func (s *SchemaService) DeleteSchema(ctx context.Context, req *catalogv1beta2.DeleteSchemaRequest) (*catalogv1beta2.DeleteSchemaResponse, error) {
// 	return nil, nil
// }

// func (s *SchemaService) GetSchema(ctx context.Context, req *catalogv1beta2.GetSchemaRequest) (*catalogv1beta2.GetSchemaResponse, error) {
// 	return nil, nil
// }

// func (s *SchemaService) ListCatalogs(ctx context.Context, req *catalogv1beta2.ListSchemasRequest) (*catalogv1beta2.ListSchemasResponse, error) {
// 	return nil, nil
// }

// func (s *SchemaService) UpdateSchema(ctx context.Context, req *catalogv1beta2.UpdateSchemaRequest) (*catalogv1beta2.UpdateSchemaResponse, error) {
// 	return nil, nil
// }
