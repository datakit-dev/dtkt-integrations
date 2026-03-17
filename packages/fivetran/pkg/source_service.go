package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	replicationv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/replication/v1beta1"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type SourceService struct {
	replicationv1beta1.UnimplementedSourceServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewSourceService(mux v1beta1.InstanceMux[*Instance]) *SourceService {
	return &SourceService{
		mux: mux,
	}
}

func (s *SourceService) ListSourceTypes(ctx context.Context, req *replicationv1beta1.ListSourceTypesRequest) (*replicationv1beta1.ListSourceTypesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSourceTypes not implemented")
}

func (s *SourceService) CheckSourceAuth(ctx context.Context, req *replicationv1beta1.CheckSourceAuthRequest) (*replicationv1beta1.CheckSourceAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckSourceAuth not implemented")
}

func (s *SourceService) GetSource(ctx context.Context, req *replicationv1beta1.GetSourceRequest) (*replicationv1beta1.GetSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSource not implemented")
}

func (s *SourceService) CreateSource(ctx context.Context, req *replicationv1beta1.CreateSourceRequest) (*replicationv1beta1.CreateSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSource not implemented")
}

func (s *SourceService) UpdateSource(ctx context.Context, req *replicationv1beta1.UpdateSourceRequest) (*replicationv1beta1.UpdateSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSource not implemented")
}

func (s *SourceService) DeleteSource(ctx context.Context, req *replicationv1beta1.DeleteSourceRequest) (*replicationv1beta1.DeleteSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSource not implemented")
}

func (s *SourceService) StartSync(ctx context.Context, req *replicationv1beta1.StartSyncRequest) (*replicationv1beta1.StartSyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartSync not implemented")
}

func (s *SourceService) StopSync(ctx context.Context, req *replicationv1beta1.StopSyncRequest) (*replicationv1beta1.StopSyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopSync not implemented")
}
