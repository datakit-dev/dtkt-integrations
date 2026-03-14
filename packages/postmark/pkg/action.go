package pkg

import (
	"context"
	"fmt"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	actionv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/action/v1beta1"

	"github.com/datakit-dev/dtkt-integrations/postmark/pkg/accountoapi"
	"github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi"

	accountoapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/accountoapi/gen"
	serveroapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi/gen"
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
		registerAccountAction(accountoapigen.CreateDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapigen.OptDomainCreationModel) (*accountoapigen.DomainExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				res, err := inst.accountClient.CreateDomain(ctx, input, accountoapigen.CreateDomainParams{
					XPostmarkAccountToken: inst.config.AccountApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.CreateSenderSignatureOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapigen.OptSenderSignatureCreationModel) (*accountoapigen.SenderSignatureExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				res, err := inst.accountClient.CreateSenderSignature(ctx, input, accountoapigen.CreateSenderSignatureParams{
					XPostmarkAccountToken: inst.config.AccountApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.SenderSignatureExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.CreateServerOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapigen.OptCreateServerPayload) (*accountoapigen.ExtendedServerInfo, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				res, err := inst.accountClient.CreateServer(ctx, input, accountoapigen.CreateServerParams{
					XPostmarkAccountToken: inst.config.AccountApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.ExtendedServerInfo:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.DeleteDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.DeleteDomainParams) (*accountoapigen.DeleteDomainOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.DeleteDomain(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DeleteDomainOK:
					return output, nil
				case *accountoapigen.DeleteDomainUnprocessableEntity:
					outputTyped := (*accountoapigen.StandardPostmarkResponse)(output)
					return nil, accountUnprocessableContentErr(outputTyped)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.DeleteSenderSignatureOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.DeleteSenderSignatureParams) (*accountoapigen.DeleteSenderSignatureOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.DeleteSenderSignature(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DeleteSenderSignatureOK:
					return output, nil
				case *accountoapigen.DeleteSenderSignatureUnprocessableEntity:
					outputTyped := (*accountoapigen.StandardPostmarkResponse)(output)
					return nil, accountUnprocessableContentErr(outputTyped)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.DeleteServerOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.DeleteServerParams) (*accountoapigen.DeleteServerOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.DeleteServer(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DeleteServerOK:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.EditDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[accountoapigen.OptDomainEditingModel, accountoapi.EditDomainParams]) (*accountoapigen.DomainExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.Params.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.EditDomain(ctx, input.Request, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.EditSenderSignatureOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[accountoapigen.OptSenderSignatureEditingModel, accountoapi.EditSenderSignatureParams]) (*accountoapigen.SenderSignatureExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.Params.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.EditSenderSignature(ctx, input.Request, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.SenderSignatureExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.EditServerInformationOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[accountoapigen.OptEditServerPayload, accountoapi.EditServerInformationParams]) (*accountoapigen.ExtendedServerInfo, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.Params.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.EditServerInformation(ctx, input.Request, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.ExtendedServerInfo:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.GetDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.GetDomainParams) (*accountoapigen.DomainExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.GetDomain(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.GetSenderSignatureOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.GetSenderSignatureParams) (*accountoapigen.SenderSignatureExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.GetSenderSignature(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.SenderSignatureExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.GetServerInformationOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.GetServerInformationParams) (*accountoapigen.ExtendedServerInfo, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.GetServerInformation(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.ExtendedServerInfo:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.ListDomainsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.ListDomainsParams) (*accountoapigen.DomainListingResults, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.ListDomains(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainListingResults:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.ListSenderSignaturesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.ListSenderSignaturesParams) (*accountoapigen.SenderListingResults, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.ListSenderSignatures(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.SenderListingResults:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.ListServersOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.ListServersParams) (*accountoapigen.ServerListingResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.ListServers(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.ServerListingResponse:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.PushTemplatesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input *accountoapigen.TemplatesPushModel) (*accountoapigen.TemplatesPushResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				res, err := inst.accountClient.PushTemplates(ctx, input, accountoapigen.PushTemplatesParams{
					XPostmarkAccountToken: inst.config.AccountApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.TemplatesPushResponse:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.RequestDkimVerificationForDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.RequestDkimVerificationForDomainParams) (*accountoapigen.DomainExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.RequestDkimVerificationForDomain(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.RequestNewDKIMKeyForSenderSignatureOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.RequestNewDKIMKeyForSenderSignatureParams) (*accountoapigen.RequestNewDKIMKeyForSenderSignatureOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.RequestNewDKIMKeyForSenderSignature(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.RequestNewDKIMKeyForSenderSignatureOK:
					return output, nil
				case *accountoapigen.RequestNewDKIMKeyForSenderSignatureUnprocessableEntity:
					outputTyped := (*accountoapigen.StandardPostmarkResponse)(output)
					return nil, accountUnprocessableContentErr(outputTyped)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.RequestReturnPathVerificationForDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.RequestReturnPathVerificationForDomainParams) (*accountoapigen.DomainExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.RequestReturnPathVerificationForDomain(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.RequestSPFVerificationForDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.RequestSPFVerificationForDomainParams) (*accountoapigen.DomainSPFResult, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.RequestSPFVerificationForDomain(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DomainSPFResult:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.RequestSPFVerificationForSenderSignatureOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.RequestSPFVerificationForSenderSignatureParams) (*accountoapigen.SenderSignatureExtendedInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.RequestSPFVerificationForSenderSignature(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.SenderSignatureExtendedInformation:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.ResendSenderSignatureConfirmationEmailOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.ResendSenderSignatureConfirmationEmailParams) (*accountoapigen.ResendSenderSignatureConfirmationEmailOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.ResendSenderSignatureConfirmationEmail(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.ResendSenderSignatureConfirmationEmailOK:
					return output, nil
				case *accountoapigen.ResendSenderSignatureConfirmationEmailUnprocessableEntity:
					outputTyped := (*accountoapigen.StandardPostmarkResponse)(output)
					return nil, accountUnprocessableContentErr(outputTyped)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerAccountAction(accountoapigen.RotateDKIMKeyForDomainOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input accountoapi.RotateDKIMKeyForDomainParams) (*accountoapigen.DKIMRotationResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.accountClient == nil {
					return nil, fmt.Errorf("account client required")
				}

				params := input.WithApiToken(inst.config.AccountApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.accountClient.RotateDKIMKeyForDomain(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *accountoapigen.DKIMRotationResponse:
					return output, nil
				case *accountoapigen.StandardPostmarkResponse:
					return nil, accountUnprocessableContentErr(output)
				case *accountoapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.ActivateBounceOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.ActivateBounceParams) (*serveroapigen.BounceActivationResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.ActivateBounce(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.BounceActivationResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.BypassRulesForInboundMessageOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.BypassRulesForInboundMessageParams) (*serveroapigen.BypassRulesForInboundMessageOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.BypassRulesForInboundMessage(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.BypassRulesForInboundMessageOK:
					return output, nil
				case *serveroapigen.BypassRulesForInboundMessageUnprocessableEntity:
					outputTyped := (*serveroapigen.StandardPostmarkResponse)(output)
					return nil, serverUnprocessableContentErr(outputTyped)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.CreateInboundRuleOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapigen.OptCreateInboundRuleRequest) (*serveroapigen.CreateInboundRuleOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.CreateInboundRule(ctx, input, serveroapigen.CreateInboundRuleParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.CreateInboundRuleOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.CreateTemplateOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input *serveroapigen.CreateTemplateRequest) (*serveroapigen.TemplateRecordResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.CreateTemplate(ctx, input, serveroapigen.CreateTemplateParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.TemplateRecordResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.DeleteInboundRuleOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.DeleteInboundRuleParams) (*serveroapigen.DeleteInboundRuleOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.DeleteInboundRule(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.DeleteInboundRuleOK:
					return output, nil
				case *serveroapigen.DeleteInboundRuleUnprocessableEntity:
					outputTyped := (*serveroapigen.StandardPostmarkResponse)(output)
					return nil, serverUnprocessableContentErr(outputTyped)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.DeleteTemplateOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.DeleteTemplateParams) (*serveroapigen.TemplateDetailResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.DeleteTemplate(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.TemplateDetailResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.EditCurrentServerConfigurationOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapigen.OptEditServerConfigurationRequest) (*serveroapigen.ServerConfigurationResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.EditCurrentServerConfiguration(ctx, input, serveroapigen.EditCurrentServerConfigurationParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.ServerConfigurationResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetBounceCountsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetBounceCountsParams) (*serveroapigen.GetBounceCountsOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetBounceCounts(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetBounceCountsOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetBounceDumpOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetBounceDumpParams) (*serveroapigen.BounceDumpResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetBounceDump(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.BounceDumpResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetBouncesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetBouncesParams) (*serveroapigen.BounceSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetBounces(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.BounceSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetClicksForSingleOutboundMessageOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetClicksForSingleOutboundMessageParams) (*serveroapigen.MessageClickSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetClicksForSingleOutboundMessage(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.MessageClickSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetCurrentServerConfigurationOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputEmpty) (*serveroapigen.ServerConfigurationResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.GetCurrentServerConfiguration(ctx, serveroapigen.GetCurrentServerConfigurationParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.ServerConfigurationResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetDeliveryStatsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputEmpty) (*serveroapigen.DeliveryStatsResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.GetDeliveryStats(ctx, serveroapigen.GetDeliveryStatsParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.DeliveryStatsResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetInboundMessageDetailsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetInboundMessageDetailsParams) (*serveroapigen.InboundMessageFullDetailsResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetInboundMessageDetails(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.InboundMessageFullDetailsResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetOpensForSingleOutboundMessageOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOpensForSingleOutboundMessageParams) (*serveroapigen.MessageOpenSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOpensForSingleOutboundMessage(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.MessageOpenSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		// registerServerAction(serveroapigen.GetOutboundClickCountsOperation,
		// 	func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundClickCountsParams) (any, error) {
		// 		params := input.WithApiToken(apiToken)

		// 		if err := validateParams(params); err != nil {
		// 			return nil, internalServerErrFromErr(err)
		// 		}

		// 		res, err := client.GetOutboundClickCounts(ctx, params)
		// 		if err != nil {
		// 			return nil, internalServerErrFromErr(err)
		// 		}

		// 		switch output := res.(type) {
		// 		case *serveroapigen.DynamicResponse:
		// 			return parseDynamicResponse(output)
		// 		case *serveroapigen.StandardPostmarkResponse:
		// 			return nil, serverUnprocessableContentErr(output)
		// 		case *serveroapigen.R500:
		// 			return nil, internalServerErr
		// 		default:
		// 			return nil, unexpectedErr(output)
		// 		}
		// 	}),
		registerServerAction(serveroapigen.GetOutboundClickCountsByBrowserFamilyOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundClickCountsByBrowserFamilyParams) (*serveroapigen.GetOutboundClickCountsByBrowserFamilyOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundClickCountsByBrowserFamily(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetOutboundClickCountsByBrowserFamilyOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		// registerServerAction(serveroapigen.GetOutboundClickCountsByLocationOperation,
		// 	func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundClickCountsByLocationParams) (any, error) {
		//    params := input.WithApiToken(apiToken)

		// 		if err := validateParams(params); err != nil {
		// 			return nil, internalServerErrFromErr(err)
		// 		}

		// 		res, err := client.GetOutboundClickCountsByLocation(ctx, params)
		// 		if err != nil {
		// 			return nil, internalServerErrFromErr(err)
		// 		}

		// 		switch output := res.(type) {
		// 		case *serveroapigen.DynamicResponse:
		// 			return parseDynamicResponse(output)
		// 		case *serveroapigen.StandardPostmarkResponse:
		// 			return nil, serverUnprocessableContentErr(output)
		// 		case *serveroapigen.R500:
		// 			return nil, internalServerErr
		// 		default:
		// 			return nil, unexpectedErr(output)
		// 		}
		// 	}),
		// registerServerAction(serveroapigen.GetOutboundClickCountsByPlatformOperation,
		// 	func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundClickCountsByPlatformParams) (any, error) {
		//    params := input.WithApiToken(apiToken)

		// 		if err := validateParams(params); err != nil {
		// 			return nil, internalServerErrFromErr(err)
		// 		}

		// 		res, err := client.GetOutboundClickCountsByPlatform(ctx, params)
		// 		if err != nil {
		// 			return nil, internalServerErrFromErr(err)
		// 		}

		// 		switch output := res.(type) {
		// 		case *serveroapigen.DynamicResponse:
		// 			return parseDynamicResponse(output)
		// 		case *serveroapigen.StandardPostmarkResponse:
		// 			return nil, serverUnprocessableContentErr(output)
		// 		case *serveroapigen.R500:
		// 			return nil, internalServerErr
		// 		default:
		// 			return nil, unexpectedErr(output)
		// 		}
		// 	}),
		registerServerAction(serveroapigen.GetOutboundMessageDetailsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundMessageDetailsParams) (*serveroapigen.OutboundMessageDetailsResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundMessageDetails(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.OutboundMessageDetailsResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetOutboundMessageDumpOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundMessageDumpParams) (*serveroapigen.OutboundMessageDumpResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundMessageDump(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.OutboundMessageDumpResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetOutboundOpenCountsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundOpenCountsParams) (*serveroapigen.GetOutboundOpenCountsOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundOpenCounts(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetOutboundOpenCountsOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetOutboundOpenCountsByEmailClientOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundOpenCountsByEmailClientParams) (*serveroapigen.GetOutboundOpenCountsByEmailClientOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundOpenCountsByEmailClient(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetOutboundOpenCountsByEmailClientOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetOutboundOpenCountsByPlatformOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundOpenCountsByPlatformParams) (*serveroapigen.GetOutboundOpenCountsByPlatformOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundOpenCountsByPlatform(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetOutboundOpenCountsByPlatformOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetOutboundOverviewStatisticsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetOutboundOverviewStatisticsParams) (*serveroapigen.OutboundOverviewStatsResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetOutboundOverviewStatistics(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.OutboundOverviewStatsResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetSentCountsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetSentCountsParams) (*serveroapigen.SentCountsResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetSentCounts(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.SentCountsResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetSingleBounceOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetSingleBounceParams) (*serveroapigen.BounceInfoResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetSingleBounce(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.BounceInfoResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetSingleTemplateOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetSingleTemplateParams) (*serveroapigen.TemplateDetailResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetSingleTemplate(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.TemplateDetailResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetSpamComplaintsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetSpamComplaintsParams) (*serveroapigen.GetSpamComplaintsOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetSpamComplaints(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetSpamComplaintsOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.GetTrackedEmailCountsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.GetTrackedEmailCountsParams) (*serveroapigen.GetTrackedEmailCountsOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.GetTrackedEmailCounts(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.GetTrackedEmailCountsOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.ListInboundRulesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.ListInboundRulesParams) (*serveroapigen.ListInboundRulesOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.ListInboundRules(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.ListInboundRulesOK:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.ListTemplatesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.ListTemplatesParams) (*serveroapigen.TemplateListingResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.ListTemplates(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.TemplateListingResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.RetryInboundMessageProcessingOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.RetryInboundMessageProcessingParams) (*serveroapigen.RetryInboundMessageProcessingOK, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.RetryInboundMessageProcessing(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.RetryInboundMessageProcessingOK:
					return output, nil
				case *serveroapigen.RetryInboundMessageProcessingUnprocessableEntity:
					outputTyped := (*serveroapigen.StandardPostmarkResponse)(output)
					return nil, serverUnprocessableContentErr(outputTyped)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SearchClicksForOutboundMessagesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.SearchClicksForOutboundMessagesParams) (*serveroapigen.MessageClickSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.SearchClicksForOutboundMessages(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.MessageClickSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SearchInboundMessagesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.SearchInboundMessagesParams) (*serveroapigen.InboundSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.SearchInboundMessages(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.InboundSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SearchOpensForOutboundMessagesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.SearchOpensForOutboundMessagesParams) (*serveroapigen.MessageOpenSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.SearchOpensForOutboundMessages(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.MessageOpenSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SearchOutboundMessagesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapi.SearchOutboundMessagesParams) (*serveroapigen.OutboundSearchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.SearchOutboundMessages(ctx, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.OutboundSearchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SendEmailOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapigen.OptSendEmailRequest) (*serveroapigen.SendEmailResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.SendEmail(ctx, input, serveroapigen.SendEmailParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.SendEmailResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SendEmailBatchOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapigen.SendEmailBatchRequest) (*serveroapigen.SendEmailBatchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.SendEmailBatch(ctx, input, serveroapigen.SendEmailBatchParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.SendEmailBatchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SendEmailBatchWithTemplatesOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input *serveroapigen.SendEmailTemplatedBatchRequest) (*serveroapigen.SendEmailBatchResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.SendEmailBatchWithTemplates(ctx, input, serveroapigen.SendEmailBatchWithTemplatesParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.SendEmailBatchResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.SendEmailWithTemplateOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input *serveroapigen.EmailWithTemplateRequest) (*serveroapigen.SendEmailResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.SendEmailWithTemplate(ctx, input, serveroapigen.SendEmailWithTemplateParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.SendEmailResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.TestTemplateContentOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input serveroapigen.OptTemplateValidationRequest) (*serveroapigen.TemplateValidationResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				res, err := inst.serverClient.TestTemplateContent(ctx, input, serveroapigen.TestTemplateContentParams{
					XPostmarkServerToken: inst.config.ServerApiToken,
				})
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.TemplateValidationResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
		registerServerAction(serveroapigen.UpdateTemplateOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[*serveroapigen.EditTemplateRequest, serveroapi.UpdateTemplateParams]) (*serveroapigen.TemplateRecordResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				} else if inst.serverClient == nil {
					return nil, fmt.Errorf("server client required")
				}

				params := input.Params.WithApiToken(inst.config.ServerApiToken)

				if err := validateParams(params); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.serverClient.UpdateTemplate(ctx, input.Request, params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *serveroapigen.TemplateRecordResponse:
					return output, nil
				case *serveroapigen.StandardPostmarkResponse:
					return nil, serverUnprocessableContentErr(output)
				case *serveroapigen.R500:
					return nil, internalServerErr
				default:
					return nil, unexpectedErr(output)
				}
			}),
	}
}

func registerAccountAction[Input, Output any](
	operation accountoapigen.OperationName,
	execFunc v1beta1.ExecuteActionFunc[*Instance, Input, Output],
) v1beta1.RegisterActionFunc[*Instance] {
	return v1beta1.NewAction(
		operation,
		accountDesc(operation),
		execFunc,
	)
}

func registerServerAction[Input, Output any](
	operation serveroapigen.OperationName,
	execFunc v1beta1.ExecuteActionFunc[*Instance, Input, Output],
) v1beta1.RegisterActionFunc[*Instance] {
	return v1beta1.NewAction(
		operation,
		serverDesc(operation),
		execFunc,
	)
}
