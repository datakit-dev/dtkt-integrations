package v1beta2

// import (
// 	"context"

// 	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
// 	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/fivetran"
// 	eventv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/event/v1beta1"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// type EventService struct {
// 	eventv1beta1.UnimplementedEventServiceServer
// 	mux      v1beta1.InstanceMux[*Instance]
// 	fivetran *fivetran.DestinationService[*Instance]
// }

// func NewEventService(mux v1beta1.InstanceMux[*Instance], fivetran *fivetran.DestinationService[*Instance]) *EventService {
// 	return &EventService{
// 		mux:      mux,
// 		fivetran: fivetran,
// 	}
// }

// func (s *EventService) CreateEventSource(ctx context.Context, req *eventv1beta1.CreateEventSourceRequest) (*eventv1beta1.CreateEventSourceResponse, error) {
// 	if req.GetEventSource().GetName() == fivetran.WebhookEventSourceName {
// 		return s.fivetran.CreateEventSource(ctx, req)
// 	}

// 	return nil, status.Errorf(codes.FailedPrecondition, "event source not found: %s", req.GetEventSource().GetName())
// }

// func (s *EventService) UpdateEventSource(ctx context.Context, req *eventv1beta1.UpdateEventSourceRequest) (*eventv1beta1.UpdateEventSourceResponse, error) {
// 	if req.GetEventSource().GetName() == fivetran.WebhookEventSourceName {
// 		return s.fivetran.UpdateEventSource(ctx, req)
// 	}

// 	return nil, status.Errorf(codes.FailedPrecondition, "event source not found: %s", req.GetEventSource().GetName())
// }

// func (s *EventService) DeleteEventSource(ctx context.Context, req *eventv1beta1.DeleteEventSourceRequest) (*eventv1beta1.DeleteEventSourceResponse, error) {
// 	if req.GetName() == fivetran.WebhookEventSourceName {
// 		return s.fivetran.DeleteEventSource(ctx, req)
// 	}

// 	return nil, status.Errorf(codes.FailedPrecondition, "event source not found: %s", req.GetName())
// }

// func (s *EventService) ListEvents(ctx context.Context, req *eventv1beta1.ListEventsRequest) (*eventv1beta1.ListEventsResponse, error) {
// 	return s.mux.Events().List(ctx, req)
// }

// func (s *EventService) ListEventSources(ctx context.Context, req *eventv1beta1.ListEventSourcesRequest) (*eventv1beta1.ListEventSourcesResponse, error) {
// 	return s.mux.EventSources().List(ctx, req)
// }

// func (s *EventService) StreamPullEvents(req *eventv1beta1.StreamPullEventsRequest, stream grpc.ServerStreamingServer[eventv1beta1.StreamPullEventsResponse]) error {
// 	return s.mux.EventSources().HandlePullStream(req, stream)
// }

// func (s *EventService) StreamPushEvents(stream grpc.BidiStreamingServer[eventv1beta1.StreamPushEventsRequest, eventv1beta1.StreamPushEventsResponse]) error {
// 	return s.mux.EventSources().HandlePushStream(stream)
// }
