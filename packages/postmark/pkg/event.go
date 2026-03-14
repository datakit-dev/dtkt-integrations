package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	eventv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/event/v1beta1"
	"google.golang.org/grpc"
)

type EventService struct {
	eventv1beta1.UnimplementedEventServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewEventService(mux v1beta1.InstanceMux[*Instance]) *EventService {
	return &EventService{
		mux: mux,
	}
}

func (s *EventService) ListEvents(ctx context.Context, req *eventv1beta1.ListEventsRequest) (*eventv1beta1.ListEventsResponse, error) {
	return s.mux.Events().List(ctx, req)
}

func (s *EventService) ListEventSources(ctx context.Context, req *eventv1beta1.ListEventSourcesRequest) (*eventv1beta1.ListEventSourcesResponse, error) {
	return s.mux.EventSources().List(ctx, req)
}

func (s *EventService) StreamPullEvents(req *eventv1beta1.StreamPullEventsRequest, stream grpc.ServerStreamingServer[eventv1beta1.StreamPullEventsResponse]) error {
	return s.mux.EventSources().HandlePullStream(req, stream)
}

func (s *EventService) StreamPushEvents(stream grpc.BidiStreamingServer[eventv1beta1.StreamPushEventsRequest, eventv1beta1.StreamPushEventsResponse]) error {
	return s.mux.EventSources().HandlePushStream(stream)
}
