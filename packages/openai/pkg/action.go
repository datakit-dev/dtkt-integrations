package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-integrations/openai/pkg/oapigen"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	actionv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/action/v1beta1"
)

type ActionService struct {
	actionv1beta1.UnimplementedActionServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewActionService(mux v1beta1.InstanceMux[*Instance]) *ActionService {
	return &ActionService{
		mux: mux,
	}
}

func RealtimeActions() []v1beta1.RegisterActionFunc[*Instance] {
	return []v1beta1.RegisterActionFunc[*Instance]{
		NewRealtimeAction[oapigen.RealtimeClientEventSessionUpdate]("SessionUpdate"),
		NewRealtimeAction[oapigen.RealtimeClientEventConversationItemCreate]("ConversationItemCreate"),
		NewRealtimeAction[oapigen.RealtimeClientEventConversationItemDelete]("ConversationItemDelete"),
		NewRealtimeAction[oapigen.RealtimeClientEventConversationItemTruncate]("ConversationItemTruncate"),
		NewRealtimeAction[oapigen.RealtimeClientEventInputAudioBufferAppend]("InputAudioBufferAppend"),
		NewRealtimeAction[oapigen.RealtimeClientEventInputAudioBufferClear]("InputAudioBufferClear"),
		NewRealtimeAction[oapigen.RealtimeClientEventInputAudioBufferCommit]("InputAudioBufferCommit"),
		NewRealtimeAction[oapigen.RealtimeClientEventResponseCreate]("ResponseCreate"),
		NewRealtimeAction[oapigen.RealtimeClientEventResponseCancel]("ResponseCancel"),
	}
}

// Returns the list of actions available to the user for this integration.
func (s *ActionService) ListActions(ctx context.Context, req *actionv1beta1.ListActionsRequest) (*actionv1beta1.ListActionsResponse, error) {
	return s.mux.Actions().List(ctx, req)
}

// Returns the full details of a single action, including its input/output schema.
func (s *ActionService) GetAction(ctx context.Context, req *actionv1beta1.GetActionRequest) (*actionv1beta1.GetActionResponse, error) {
	return s.mux.Actions().Get(ctx, req)
}

// Runs the specified action with the provided input data and returns the result.
func (s *ActionService) ExecuteAction(ctx context.Context, req *actionv1beta1.ExecuteActionRequest) (*actionv1beta1.ExecuteActionResponse, error) {
	return s.mux.Actions().Execute(ctx, req)
}
