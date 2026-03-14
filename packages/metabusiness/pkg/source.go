package pkg

import (
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	replicationv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/replication/v1beta1"
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
