package pkg

import (
	"github.com/datakit-dev/dtkt-integrations/openai/pkg/oapigen"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/common"
)

var (
	EventTypes = map[string]string{
		"SessionUpdate":                                    eventTypeFor("RealtimeClientEventSessionUpdate"),
		"ResponseCreate":                                   eventTypeFor("RealtimeClientEventResponseCreate"),
		"ResponseCancel":                                   eventTypeFor("RealtimeClientEventResponseCancel"),
		"ConversationItemCreate":                           eventTypeFor("RealtimeClientEventConversationItemCreate"),
		"ConversationItemDelete":                           eventTypeFor("RealtimeClientEventConversationItemDelete"),
		"ConversationItemTruncate":                         eventTypeFor("RealtimeClientEventConversationItemTruncate"),
		"InputAudioBufferAppend":                           eventTypeFor("RealtimeClientEventInputAudioBufferAppend"),
		"InputAudioBufferClear":                            eventTypeFor("RealtimeClientEventInputAudioBufferClear"),
		"InputAudioBufferCommit":                           eventTypeFor("RealtimeClientEventInputAudioBufferCommit"),
		"ConversationCreated":                              eventTypeFor("RealtimeServerEventConversationCreated"),
		"ConversationItemCreated":                          eventTypeFor("RealtimeServerEventConversationItemCreated"),
		"ConversationItemDeleted":                          eventTypeFor("RealtimeServerEventConversationItemDeleted"),
		"ConversationItemInputAudioTranscriptionCompleted": eventTypeFor("RealtimeServerEventConversationItemInputAudioTranscriptionCompleted"),
		"ConversationItemInputAudioTranscriptionFailed":    eventTypeFor("RealtimeServerEventConversationItemInputAudioTranscriptionFailed"),
		"ConversationItemTruncated":                        eventTypeFor("RealtimeServerEventConversationItemTruncated"),
		"SessionCreated":                                   eventTypeFor("RealtimeServerEventSessionCreated"),
		"SessionUpdated":                                   eventTypeFor("RealtimeServerEventSessionUpdated"),
		"Error":                                            eventTypeFor("RealtimeServerEventError"),
		"InputAudioBufferCleared":                          eventTypeFor("RealtimeServerEventInputAudioBufferCleared"),
		"InputAudioBufferCommitted":                        eventTypeFor("RealtimeServerEventInputAudioBufferCommitted"),
		"InputAudioBufferSpeechStarted":                    eventTypeFor("RealtimeServerEventInputAudioBufferSpeechStarted"),
		"InputAudioBufferSpeechStopped":                    eventTypeFor("RealtimeServerEventInputAudioBufferSpeechStopped"),
		"RateLimitsUpdated":                                eventTypeFor("RealtimeServerEventRateLimitsUpdated"),
		"ResponseAudioDelta":                               eventTypeFor("RealtimeServerEventResponseAudioDelta"),
		"ResponseAudioDone":                                eventTypeFor("RealtimeServerEventResponseAudioDone"),
		"ResponseAudioTranscriptDelta":                     eventTypeFor("RealtimeServerEventResponseAudioTranscriptDelta"),
		"ResponseAudioTranscriptDone":                      eventTypeFor("RealtimeServerEventResponseAudioTranscriptDone"),
		"ResponseContentPartAdded":                         eventTypeFor("RealtimeServerEventResponseContentPartAdded"),
		"ResponseContentPartDone":                          eventTypeFor("RealtimeServerEventResponseContentPartDone"),
		"ResponseCreated":                                  eventTypeFor("RealtimeServerEventResponseCreated"),
		"ResponseDone":                                     eventTypeFor("RealtimeServerEventResponseDone"),
		"ResponseFunctionCallArgumentsDelta":               eventTypeFor("RealtimeServerEventResponseFunctionCallArgumentsDelta"),
		"ResponseFunctionCallArgumentsDone":                eventTypeFor("RealtimeServerEventResponseFunctionCallArgumentsDone"),
		"ResponseOutputItemAdded":                          eventTypeFor("RealtimeServerEventResponseOutputItemAdded"),
		"ResponseOutputItemDone":                           eventTypeFor("RealtimeServerEventResponseOutputItemDone"),
		"ResponseTextDelta":                                eventTypeFor("RealtimeServerEventResponseTextDelta"),
		"ResponseTextDone":                                 eventTypeFor("RealtimeServerEventResponseTextDone"),
	}
	EventDescriptions = map[string]string{
		"SessionUpdate":                                    eventDescriptionFor("RealtimeClientEventSessionUpdate"),
		"ResponseCreate":                                   eventDescriptionFor("RealtimeClientEventResponseCreate"),
		"ResponseCancel":                                   eventDescriptionFor("RealtimeClientEventResponseCancel"),
		"ConversationItemCreate":                           eventDescriptionFor("RealtimeClientEventConversationItemCreate"),
		"ConversationItemDelete":                           eventDescriptionFor("RealtimeClientEventConversationItemDelete"),
		"ConversationItemTruncate":                         eventDescriptionFor("RealtimeClientEventConversationItemTruncate"),
		"InputAudioBufferAppend":                           eventDescriptionFor("RealtimeClientEventInputAudioBufferAppend"),
		"InputAudioBufferClear":                            eventDescriptionFor("RealtimeClientEventInputAudioBufferClear"),
		"InputAudioBufferCommit":                           eventDescriptionFor("RealtimeClientEventInputAudioBufferCommit"),
		"ConversationCreated":                              eventDescriptionFor("RealtimeServerEventConversationCreated"),
		"ConversationItemCreated":                          eventDescriptionFor("RealtimeServerEventConversationItemCreated"),
		"ConversationItemDeleted":                          eventDescriptionFor("RealtimeServerEventConversationItemDeleted"),
		"ConversationItemInputAudioTranscriptionCompleted": eventDescriptionFor("RealtimeServerEventConversationItemInputAudioTranscriptionCompleted"),
		"ConversationItemInputAudioTranscriptionFailed":    eventDescriptionFor("RealtimeServerEventConversationItemInputAudioTranscriptionFailed"),
		"ConversationItemTruncated":                        eventDescriptionFor("RealtimeServerEventConversationItemTruncated"),
		"SessionCreated":                                   eventDescriptionFor("RealtimeServerEventSessionCreated"),
		"SessionUpdated":                                   eventDescriptionFor("RealtimeServerEventSessionUpdated"),
		"Error":                                            eventDescriptionFor("RealtimeServerEventError"),
		"InputAudioBufferCleared":                          eventDescriptionFor("RealtimeServerEventInputAudioBufferCleared"),
		"InputAudioBufferCommitted":                        eventDescriptionFor("RealtimeServerEventInputAudioBufferCommitted"),
		"InputAudioBufferSpeechStarted":                    eventDescriptionFor("RealtimeServerEventInputAudioBufferSpeechStarted"),
		"InputAudioBufferSpeechStopped":                    eventDescriptionFor("RealtimeServerEventInputAudioBufferSpeechStopped"),
		"RateLimitsUpdated":                                eventDescriptionFor("RealtimeServerEventRateLimitsUpdated"),
		"ResponseAudioDelta":                               eventDescriptionFor("RealtimeServerEventResponseAudioDelta"),
		"ResponseAudioDone":                                eventDescriptionFor("RealtimeServerEventResponseAudioDone"),
		"ResponseAudioTranscriptDelta":                     eventDescriptionFor("RealtimeServerEventResponseAudioTranscriptDelta"),
		"ResponseAudioTranscriptDone":                      eventDescriptionFor("RealtimeServerEventResponseAudioTranscriptDone"),
		"ResponseContentPartAdded":                         eventDescriptionFor("RealtimeServerEventResponseContentPartAdded"),
		"ResponseContentPartDone":                          eventDescriptionFor("RealtimeServerEventResponseContentPartDone"),
		"ResponseCreated":                                  eventDescriptionFor("RealtimeServerEventResponseCreated"),
		"ResponseDone":                                     eventDescriptionFor("RealtimeServerEventResponseDone"),
		"ResponseFunctionCallArgumentsDelta":               eventDescriptionFor("RealtimeServerEventResponseFunctionCallArgumentsDelta"),
		"ResponseFunctionCallArgumentsDone":                eventDescriptionFor("RealtimeServerEventResponseFunctionCallArgumentsDone"),
		"ResponseOutputItemAdded":                          eventDescriptionFor("RealtimeServerEventResponseOutputItemAdded"),
		"ResponseOutputItemDone":                           eventDescriptionFor("RealtimeServerEventResponseOutputItemDone"),
		"ResponseTextDelta":                                eventDescriptionFor("RealtimeServerEventResponseTextDelta"),
		"ResponseTextDone":                                 eventDescriptionFor("RealtimeServerEventResponseTextDone"),
	}
)

func eventTypeFor(name string) string {
	return common.JSONValueMust[string](oapigen.OpenAPISpec, `.components.schemas.["`+name+`"].["x-oaiMeta"].name`)
}

func eventDescriptionFor(name string) string {
	return common.JSONValueMust[string](oapigen.OpenAPISpec, `.components.schemas.["`+name+`"].description`)
}
