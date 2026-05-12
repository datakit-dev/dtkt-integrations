package pkg

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/timestamppb"

	testbinv1beta "github.com/datakit-dev/dtkt-integrations/testbin/pkg/proto/integration/testbin/v1beta"
)

type InspectService struct {
	testbinv1beta.UnimplementedInspectServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewInspectService(mux v1beta1.InstanceMux[*Instance]) *InspectService {
	return &InspectService{mux: mux}
}

func (s *InspectService) GetConfig(ctx context.Context, _ *testbinv1beta.GetConfigRequest) (*testbinv1beta.GetConfigResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}
	return &testbinv1beta.GetConfigResponse{
		Payload: inst.config.GetPayload(),
	}, nil
}

func (s *InspectService) Anything(ctx context.Context, req *testbinv1beta.AnythingRequest) (*testbinv1beta.AnythingResponse, error) {
	resp := &testbinv1beta.AnythingResponse{
		Metadata: map[string]*testbinv1beta.StringValues{},
		Payload:  req.GetPayload(),
	}

	if method, ok := grpc.Method(ctx); ok {
		resp.Method = method
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vs := range md {
			values := make([]string, len(vs))
			if strings.HasSuffix(k, "-bin") {
				for i, v := range vs {
					values[i] = base64.StdEncoding.EncodeToString([]byte(v))
				}
			} else {
				copy(values, vs)
			}
			resp.Metadata[k] = &testbinv1beta.StringValues{Values: values}
		}
	}

	if dl, ok := ctx.Deadline(); ok {
		resp.Deadline = timestamppb.New(dl)
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		resp.Peer = p.Addr.String()
	}

	return resp, nil
}
