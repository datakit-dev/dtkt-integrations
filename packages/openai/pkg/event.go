package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-integrations/openai/pkg/oapigen"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"google.golang.org/grpc"

	eventv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/event/v1beta1"
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

func RealtimeEvents() []v1beta1.RegisterEventFunc[*Instance] {
	return []v1beta1.RegisterEventFunc[*Instance]{
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventError](
			EventTypes["Error"], //"RealtimeServerEventError",
			EventDescriptions["Error"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventSessionUpdate](
			EventTypes["SessionUpdate"], //"RealtimeClientEventSessionUpdate",
			EventDescriptions["SessionUpdate"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventResponseCreate](
			EventTypes["ResponseCreate"], //"RealtimeClientEventResponseCreate",
			EventDescriptions["ResponseCreate"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventResponseCancel](
			EventTypes["ResponseCancel"], //"RealtimeClientEventResponseCancel",
			EventDescriptions["ResponseCancel"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventConversationItemCreate](
			EventTypes["ConversationItemCreate"], //"RealtimeClientEventConversationItemCreate",
			EventDescriptions["ConversationItemCreate"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventConversationItemDelete](
			EventTypes["ConversationItemDelete"], //"RealtimeClientEventConversationItemDelete",
			EventDescriptions["ConversationItemDelete"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventConversationItemTruncate](
			EventTypes["ConversationItemTruncate"], //"RealtimeClientEventConversationItemTruncate",
			EventDescriptions["ConversationItemTruncate"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventInputAudioBufferAppend](
			EventTypes["InputAudioBufferAppend"], //"RealtimeClientEventInputAudioBufferAppend",
			EventDescriptions["InputAudioBufferAppend"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventInputAudioBufferClear](
			EventTypes["InputAudioBufferClear"], //"RealtimeClientEventInputAudioBufferClear",
			EventDescriptions["InputAudioBufferClear"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeClientEventInputAudioBufferCommit](
			EventTypes["InputAudioBufferCommit"], //"RealtimeClientEventInputAudioBufferCommit",
			EventDescriptions["InputAudioBufferCommit"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventConversationCreated](
			EventTypes["ConversationCreated"], //"RealtimeServerEventConversationCreated",
			EventDescriptions["ConversationCreated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventConversationItemCreated](
			EventTypes["ConversationItemCreated"], //"RealtimeServerEventConversationItemCreated",
			EventDescriptions["ConversationItemCreated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventConversationItemDeleted](
			EventTypes["ConversationItemDeleted"], //"RealtimeServerEventConversationItemDeleted",
			EventDescriptions["ConversationItemDeleted"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventConversationItemInputAudioTranscriptionCompleted](
			EventTypes["ConversationItemInputAudioTranscriptionCompleted"], //"RealtimeServerEventConversationItemInputAudioTranscriptionCompleted",
			EventDescriptions["ConversationItemInputAudioTranscriptionCompleted"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventConversationItemInputAudioTranscriptionFailed](
			EventTypes["ConversationItemInputAudioTranscriptionFailed"], //"RealtimeServerEventConversationItemInputAudioTranscriptionFailed",
			EventDescriptions["ConversationItemInputAudioTranscriptionFailed"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventConversationItemTruncated](
			EventTypes["ConversationItemTruncated"], //"RealtimeServerEventConversationItemTruncated",
			EventDescriptions["ConversationItemTruncated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventSessionCreated](
			EventTypes["SessionCreated"], //"RealtimeServerEventSessionCreated",
			EventDescriptions["SessionCreated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventSessionUpdated](
			EventTypes["SessionUpdated"], //"RealtimeServerEventSessionUpdated",
			EventDescriptions["SessionUpdated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventInputAudioBufferCleared](
			EventTypes["InputAudioBufferCleared"], //"RealtimeServerEventInputAudioBufferCleared",
			EventDescriptions["InputAudioBufferCleared"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventInputAudioBufferCommitted](
			EventTypes["InputAudioBufferCommitted"], //"RealtimeServerEventInputAudioBufferCommitted",
			EventDescriptions["InputAudioBufferCommitted"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventInputAudioBufferSpeechStarted](
			EventTypes["InputAudioBufferSpeechStarted"], //"RealtimeServerEventInputAudioBufferSpeechStarted",
			EventDescriptions["InputAudioBufferSpeechStarted"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventInputAudioBufferSpeechStopped](
			EventTypes["InputAudioBufferSpeechStopped"], //"RealtimeServerEventInputAudioBufferSpeechStopped",
			EventDescriptions["InputAudioBufferSpeechStopped"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventRateLimitsUpdated](
			EventTypes["RateLimitsUpdated"], //"RealtimeServerEventRateLimitsUpdated",
			EventDescriptions["RateLimitsUpdated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseAudioDelta](
			EventTypes["ResponseAudioDelta"], //"RealtimeServerEventResponseAudioDelta",
			EventDescriptions["ResponseAudioDelta"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseAudioDone](
			EventTypes["ResponseAudioDone"], //"RealtimeServerEventResponseAudioDone",
			EventDescriptions["ResponseAudioDone"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseAudioTranscriptDelta](
			EventTypes["ResponseAudioTranscriptDelta"], //"RealtimeServerEventResponseAudioTranscriptDelta",
			EventDescriptions["ResponseAudioTranscriptDelta"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseAudioTranscriptDone](
			EventTypes["ResponseAudioTranscriptDone"], //"RealtimeServerEventResponseAudioTranscriptDone",
			EventDescriptions["ResponseAudioTranscriptDone"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseContentPartAdded](
			EventTypes["ResponseContentPartAdded"], //"RealtimeServerEventResponseContentPartAdded",
			EventDescriptions["ResponseContentPartAdded"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseContentPartDone](
			EventTypes["ResponseContentPartDone"], //"RealtimeServerEventResponseContentPartDone",
			EventDescriptions["ResponseContentPartDone"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseCreated](
			EventTypes["ResponseCreated"], //"RealtimeServerEventResponseCreated",
			EventDescriptions["ResponseCreated"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseDone](
			EventTypes["ResponseDone"], //"RealtimeServerEventResponseDone",
			EventDescriptions["ResponseDone"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseFunctionCallArgumentsDelta](
			EventTypes["ResponseFunctionCallArgumentsDelta"], //"RealtimeServerEventResponseFunctionCallArgumentsDelta",
			EventDescriptions["ResponseFunctionCallArgumentsDelta"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseFunctionCallArgumentsDone](
			EventTypes["ResponseFunctionCallArgumentsDone"], //"RealtimeServerEventResponseFunctionCallArgumentsDone",
			EventDescriptions["ResponseFunctionCallArgumentsDone"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseOutputItemAdded](
			EventTypes["ResponseOutputItemAdded"], //"RealtimeServerEventResponseOutputItemAdded",
			EventDescriptions["ResponseOutputItemAdded"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseOutputItemDone](
			EventTypes["ResponseOutputItemDone"], //"RealtimeServerEventResponseOutputItemDone",
			EventDescriptions["ResponseOutputItemDone"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseTextDelta](
			EventTypes["ResponseTextDelta"], //"RealtimeServerEventResponseTextDelta",
			EventDescriptions["ResponseTextDelta"],
		),
		v1beta1.RegisterEvent[*Instance, *oapigen.RealtimeServerEventResponseTextDone](
			EventTypes["ResponseTextDone"], //"RealtimeServerEventResponseTextDone",
			EventDescriptions["ResponseTextDone"],
		),
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
