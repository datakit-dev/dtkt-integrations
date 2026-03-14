package v1beta2

import (
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta2 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta2"
)

type QueryService struct {
	catalogv1beta2.UnimplementedQueryServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewQueryService(mux v1beta1.InstanceMux[*Instance]) *QueryService {
	return &QueryService{
		mux: mux,
	}
}

// func (s *QueryService) ValidateQuery(context.Context, *catalogv1beta2.ValidateQueryRequest) (*catalogv1beta2.ValidateQueryResponse, error) {
// 	return nil, nil
// }

// func (s *QueryService) ListQueryResults(context.Context, *catalogv1beta2.ListQueryResultsRequest) (*catalogv1beta2.ListQueryResultsResponse, error) {
// 	return nil, nil
// }

// func (s *QueryService) StreamQueryResults(*catalogv1beta2.StreamQueryResultsRequest, grpc.ServerStreamingServer[catalogv1beta2.StreamQueryResultsResponse]) error {
// 	return nil
// }
