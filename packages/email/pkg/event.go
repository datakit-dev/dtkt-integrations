package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/datakit-dev/dtkt-integrations/email/pkg/imap"
	"github.com/datakit-dev/dtkt-integrations/email/pkg/pop3"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
	eventv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/event/v1beta1"
	"google.golang.org/grpc"
)

type (
	EventService struct {
		eventv1beta1.UnimplementedEventServiceServer
		mux v1beta1.InstanceMux[*Instance]
	}
	EventType    string
	SourceConfig struct {
		LastEventId int `json:"last_event_id,omitempty"` // Last processed event ID
	}
	SourceHandler struct {
		config      *SourceConfig
		eventCh     chan *v1beta1.EventWithPayload
		pop3Service *pop3.Pop3Service
		imapService *imap.ImapService
	}
)

func NewEventService(mux v1beta1.InstanceMux[*Instance]) *EventService {
	return &EventService{
		mux: mux,
	}
}

// Retrieves a list of events supported by service.
func (s *EventService) ListEvents(ctx context.Context, req *eventv1beta1.ListEventsRequest) (*eventv1beta1.ListEventsResponse, error) {
	return s.mux.Events().List(ctx, req)
}

// Retrieves a list of event sources supported by service.
func (s *EventService) ListEventSources(ctx context.Context, req *eventv1beta1.ListEventSourcesRequest) (*eventv1beta1.ListEventSourcesResponse, error) {
	return s.mux.EventSources().List(ctx, req)
}

// Create a server side stream of events for a PULL event source (e.g., cron jobs).
func (s *EventService) StreamPullEvents(req *eventv1beta1.StreamPullEventsRequest, stream grpc.ServerStreamingServer[eventv1beta1.StreamPullEventsResponse]) error {
	return s.mux.EventSources().HandlePullStream(req, stream)
}

// Compatibility check for SourceHandler to ensure it implements the PullSource interface
var _ v1beta1.PullSourceHandler[*SourceConfig] = (*SourceHandler)(nil)

const (
	EmailReceivedEvent EventType = "EmailReceived"

	MaxEvents = 100 // Maximum number of events to fetch in one go
)

func (m EventType) String() string {
	return string(m)
}

func EmailEventSource() v1beta1.RegisterSourceFunc[*Instance] {
	return v1beta1.NewPullSource(
		"Emails",
		"Provides a stream of received emails.",
		false, true,
		0,
		func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], config *SourceConfig) (*SourceHandler, error) {
			inst, err := mux.GetInstance(ctx)
			if err != nil {
				return nil, err
			}

			sourceHandler := NewSourceHandler()
			if inst.config.Pop3 != nil {
				pop3Service, err := pop3.NewPop3Service(inst.config.Pop3)
				if err != nil {
					return nil, err
				}

				err = pop3Service.CheckConfig()
				if err != nil {
					return nil, err
				}

				sourceHandler.SetPop3Service(pop3Service)
			} else {
				return nil, fmt.Errorf("pop3 or imap service required")

				// } else if inst.config.Imap != nil {
				// 	imapService, err := imap.NewImapService(inst.config.Imap, mux.Events())
				// 	if err != nil {
				// 		return nil, err
				// 	}

				// 	err = imapService.CheckConfig()
				// 	if err != nil {
				// 		return nil, err
				// 	}

				// 	sourceHandler.SetImapService(imapService)
				// } else {
				// 	return nil, fmt.Errorf("pop3 or imap service required")
			}

			return sourceHandler, nil
		},
	)
}

var Events = []v1beta1.RegisterEventFunc[*Instance]{
	v1beta1.RegisterEvent[*Instance, *emailv1beta1.ReceivedEmail](
		EmailReceivedEvent.String(),
		"Event triggered when a new email is received.",
	),
}

func NewSourceHandler() *SourceHandler {
	sourceHandler := &SourceHandler{
		config:  &SourceConfig{},
		eventCh: make(chan *v1beta1.EventWithPayload, MaxEvents),
	}

	return sourceHandler
}

func (h *SourceHandler) PullEvent(ctx context.Context, events *v1beta1.EventRegistry) (*v1beta1.EventWithPayload, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case event, ok := <-h.eventCh:
			if !ok {
				// Channel closed and drained
				return nil, nil
			}
			return event, nil

		default:
			hasEvents := false

			if h.pop3Service != nil {
				events, lastEventId, err := h.pop3Service.GetEvents(ctx, events, h.config.LastEventId+1, MaxEvents)
				h.SetLastEventId(lastEventId)

				if err != nil {
					return nil, fmt.Errorf("failed to get event source response(s) from POP3 service: %v", err)
				}

				hasEvents = len(events) > 0
				for _, resp := range events {
					h.eventCh <- resp
				}
			}

			if !hasEvents {
				// If no events sleep for 5 minutes before checking again
				time.Sleep(5 * time.Minute)
			}

			return nil, nil
		}
	}
}

func (h *SourceHandler) ConfigEqual(config *SourceConfig) bool {
	return h.config.LastEventId == config.LastEventId
}

func (h *SourceHandler) SetConfig(ctx context.Context, config *SourceConfig) (*SourceHandler, error) {
	h.config = config
	return h, nil
}

func (h *SourceHandler) GetLastEventId() int {
	if h.config != nil {
		return h.config.LastEventId
	}

	return 0
}

func (h *SourceHandler) SetLastEventId(lastEventId int) *SourceHandler {
	h.config.LastEventId = lastEventId
	return h
}

func (h *SourceHandler) SetPop3Service(pop3Service *pop3.Pop3Service) *SourceHandler {
	h.pop3Service = pop3Service
	return h
}

func (h *SourceHandler) SetImapService(imapService *imap.ImapService) *SourceHandler {
	h.imapService = imapService
	return h
}

func (h *SourceHandler) Close() error {
	return nil
}
