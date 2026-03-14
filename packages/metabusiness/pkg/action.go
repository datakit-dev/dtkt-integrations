package pkg

import (
	"context"
	"net/http"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	actionv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/action/v1beta1"
)

type ActionService struct {
	actionv1beta1.ActionServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewActionService(mux v1beta1.InstanceMux[*Instance]) *ActionService {
	return &ActionService{
		mux: mux,
	}
}

func (s *ActionService) ExecuteAction(ctx context.Context, req *actionv1beta1.ExecuteActionRequest) (*actionv1beta1.ExecuteActionResponse, error) {
	return s.mux.Actions().Execute(ctx, req)
}

func (s *ActionService) ListActions(ctx context.Context, req *actionv1beta1.ListActionsRequest) (*actionv1beta1.ListActionsResponse, error) {
	return s.mux.Actions().List(ctx, req)
}

func (s *ActionService) GetAction(ctx context.Context, req *actionv1beta1.GetActionRequest) (*actionv1beta1.GetActionResponse, error) {
	return s.mux.Actions().Get(ctx, req)
}

func Actions() []v1beta1.RegisterActionFunc[*Instance] {
	return []v1beta1.RegisterActionFunc[*Instance]{
		v1beta1.NewAction(http.MethodGet, "Perform a GET request to Meta Graph API",
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input Input) (Output, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}

				return inst.client.Get(ctx, input)
			},
		),
		v1beta1.NewAction(http.MethodPost, "Perform a POST request to Meta Graph API",
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input Input) (Output, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}

				return inst.client.Post(ctx, input)
			},
		),
		v1beta1.NewAction(http.MethodPut, "Perform a PUT request to Meta Graph API",
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input Input) (Output, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}

				return inst.client.Put(ctx, input)
			},
		),
		v1beta1.NewAction(http.MethodDelete, "Perform a DELETE request to Meta Graph API",
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input Input) (Output, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}

				return inst.client.Delete(ctx, input)
			},
		),
	}
}
