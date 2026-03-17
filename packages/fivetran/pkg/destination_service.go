package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	replicationv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/replication/v1beta1"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type DestinationService struct {
	replicationv1beta1.UnimplementedDestinationServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewDestinationService(mux v1beta1.InstanceMux[*Instance]) *DestinationService {
	return &DestinationService{
		mux: mux,
	}
}

func (s *DestinationService) ListDestinationTypes(ctx context.Context, req *replicationv1beta1.ListDestinationTypesRequest) (*replicationv1beta1.ListDestinationTypesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDestinationTypes not implemented")
}

func (s *DestinationService) GetDestination(ctx context.Context, req *replicationv1beta1.GetDestinationRequest) (*replicationv1beta1.GetDestinationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDestination not implemented")
}

func (s *DestinationService) CreateDestination(ctx context.Context, req *replicationv1beta1.CreateDestinationRequest) (*replicationv1beta1.CreateDestinationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDestination not implemented")
}

func (s *DestinationService) UpdateDestination(ctx context.Context, req *replicationv1beta1.UpdateDestinationRequest) (*replicationv1beta1.UpdateDestinationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDestination not implemented")
}

func (s *DestinationService) DeleteDestination(ctx context.Context, req *replicationv1beta1.DeleteDestinationRequest) (*replicationv1beta1.DeleteDestinationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteDestination not implemented")
}
