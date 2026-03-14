package v1beta2

import (
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta2 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta2"
)

type TableService struct {
	catalogv1beta2.UnimplementedTableServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewTableService(mux v1beta1.InstanceMux[*Instance]) *TableService {
	return &TableService{
		mux: mux,
	}
}

// func (s *TableService) WriteRows(grpc.ClientStreamingServer[catalogv1beta2.WriteRowsRequest, catalogv1beta2.WriteRowsResponse]) error {
// 	return nil
// }

// func (s *TableService) ReadRows(*catalogv1beta2.ReadRowsRequest, grpc.ServerStreamingServer[catalogv1beta2.ReadRowsResponse]) error {
// 	return nil
// }

// func (s *TableService) CreateTable(ctx context.Context, req *catalogv1beta2.CreateTableRequest) (*catalogv1beta2.CreateTableResponse, error) {
// 	return nil, nil
// }

// func (s *TableService) DeleteTable(ctx context.Context, req *catalogv1beta2.DeleteTableRequest) (*catalogv1beta2.DeleteTableResponse, error) {
// 	return nil, nil
// }

// func (s *TableService) GetTable(ctx context.Context, req *catalogv1beta2.GetTableRequest) (*catalogv1beta2.GetTableResponse, error) {
// 	return nil, nil
// }

// func (s *TableService) ListCatalogs(ctx context.Context, req *catalogv1beta2.ListTablesRequest) (*catalogv1beta2.ListTablesResponse, error) {
// 	return nil, nil
// }

// func (s *TableService) UpdateTable(ctx context.Context, req *catalogv1beta2.UpdateTableRequest) (*catalogv1beta2.UpdateTableResponse, error) {
// 	return nil, nil
// }
