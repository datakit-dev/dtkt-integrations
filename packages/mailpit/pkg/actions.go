package pkg

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	actionv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/action/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datakit-dev/dtkt-integrations/mailpit/pkg/oapi"
	oapigen "github.com/datakit-dev/dtkt-integrations/mailpit/pkg/oapi/gen"
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

func descExpr(operation string) string {
	return fmt.Sprintf(`.paths[] | .[] | select((.operationId|ascii_upcase) == ("%s"|ascii_upcase))`, operation)
}

func desc(operation string) string {
	expr := descExpr(operation)
	summary, _ := oapi.OpenAPISpec.Value(expr + " | .summary")
	desc, _ := oapi.OpenAPISpec.Value(expr + " | .description")

	return fmt.Sprintf("%s.\n\n%s\n", summary, desc)
}

func Actions() []v1beta1.RegisterActionFunc[*Instance] {
	return []v1beta1.RegisterActionFunc[*Instance]{
		registerAction(oapigen.GetChaosOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputEmpty) (*oapigen.ChaosTriggers, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}

				res, err := inst.client.GetChaos(ctx)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ChaosTriggers:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.SetChaosParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.OptChaosTriggers) (*oapigen.ChaosTriggers, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}

				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.SetChaosParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ChaosTriggers:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.AppInformationOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], _ inputEmpty) (*oapigen.AppInformation, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				res, err := inst.client.AppInformation(ctx)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.AppInformation:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetMessageParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.GetMessageParamsParams) (*oapigen.Message, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.GetMessageParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.Message:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetHeadersParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.GetHeadersParamsParams) (*oapigen.MessageHeadersResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.GetHeadersParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.MessageHeadersResponse:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.HTMLCheckParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.HTMLCheckParamsParams) (*oapigen.HTMLCheckResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.HTMLCheckParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.HTMLCheckResponse:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.LinkCheckParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.LinkCheckParamsParams) (*oapigen.LinkCheckResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.LinkCheckParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.LinkCheckResponse:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.AttachmentParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.AttachmentParamsParams) (*binaryOutput, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.AttachmentParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.BinaryResponseHeaders:
					res := &binaryOutput{
						ContentType: output.ContentType,
					}
					_, err = output.GetResponse().Read(res.Bytes)
					if err != nil {
						return nil, internalServerErrFromErr(err)
					}
					return res, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.ThumbnailParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.ThumbnailParamsParams) (*binaryOutput, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.ThumbnailParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.BinaryResponseHeaders:
					res := &binaryOutput{
						ContentType: output.ContentType,
					}
					_, err = output.GetResponse().Read(res.Bytes)
					if err != nil {
						return nil, internalServerErrFromErr(err)
					}
					return res, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.DownloadRawParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.DownloadRawParamsParams) (*oapigen.DownloadRawParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.DownloadRawParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.DownloadRawParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.ReleaseMessageParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[oapigen.OptReleaseMessageParamsReq, oapigen.ReleaseMessageParamsParams]) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.ReleaseMessageParams(ctx, input.Request, input.Params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.SpamAssassinCheckParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.SpamAssassinCheckParamsParams) (*oapigen.SpamAssassinResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.SpamAssassinCheckParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.SpamAssassinResponse:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.DeleteMessagesParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.OptDeleteMessagesParamsReq) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.DeleteMessagesParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetMessagesParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.GetMessagesParamsParams) (*oapigen.MessagesSummary, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.GetMessagesParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.MessagesSummary:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.SetReadStatusParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[*oapigen.OptSetReadStatusParamsReq, oapigen.SetReadStatusParamsParams]) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.SetReadStatusParams(ctx, *input.Request, input.Params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.DeleteSearchParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.DeleteSearchParamsParams) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.DeleteSearchParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.SearchParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.SearchParamsParams) (*oapigen.MessagesSummary, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.SearchParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.MessagesSummary:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.SendMessageParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.OptSendMessageParamsReq) (*oapigen.SendMessageResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.SendMessageParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.SendMessageResponse:
					return output, nil
				case *oapigen.JSONErrorResponse:
					return nil, status.Errorf(codes.Internal, "send message error: %s", output.Error.Value)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetAllTagsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], _ inputEmpty) (*oapigen.GetAllTagsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				res, err := inst.client.GetAllTags(ctx)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.GetAllTagsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.SetTagsParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.OptSetTagsParamsReq) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.SetTagsParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.DeleteTagParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.DeleteTagParamsParams) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.DeleteTagParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.RenameTagParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input inputRequestParams[oapigen.OptRenameTagParamsReq, oapigen.RenameTagParamsParams]) (*oapigen.ReleaseMessageParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.RenameTagParams(ctx, input.Request, input.Params)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.ReleaseMessageParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetWebUIConfigurationOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], _ inputEmpty) (*oapigen.WebUIConfigurationResponse, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				res, err := inst.client.GetWebUIConfiguration(ctx)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.WebUIConfigurationResponse:
					return output, nil
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetMessageHTMLParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.GetMessageHTMLParamsParams) (*oapigen.GetMessageHTMLParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.GetMessageHTMLParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.GetMessageHTMLParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),

		registerAction(oapigen.GetMessageTextParamsOperation,
			func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input oapigen.GetMessageTextParamsParams) (*oapigen.DownloadRawParamsOKApplicationJSON, error) {
				inst, err := mux.GetInstance(ctx)
				if err != nil {
					return nil, err
				}
				if err := validateParams(input); err != nil {
					return nil, internalServerErrFromErr(err)
				}

				res, err := inst.client.GetMessageTextParams(ctx, input)
				if err != nil {
					return nil, internalServerErrFromErr(err)
				}

				switch output := res.(type) {
				case *oapigen.DownloadRawParamsOKApplicationJSON:
					return output, nil
				case *oapigen.GetMessageParamsNotFoundApplicationJSON:
					return nil, paramsNotFoundErr(output)
				case *oapigen.GetChaosBadRequestApplicationJSON:
					return nil, chaosErr(output)
				default:
					return nil, unexpectedErr(output)
				}
			}),
	}
}

type inputRequestParams[Request any, Params any] struct {
	Request Request
	Params  Params
}

type inputEmpty struct{}

type binaryOutput struct {
	ContentType string
	Bytes       []byte
}

func registerAction[Input, Output any](
	operation oapigen.OperationName,
	execFunc v1beta1.ExecuteActionFunc[*Instance, Input, Output],
) v1beta1.RegisterActionFunc[*Instance] {
	name, _ := strings.CutSuffix(string(operation), "Params")
	return v1beta1.NewAction(
		name,
		desc(operation),
		execFunc,
	)
}

func internalServerErrFromErr(err error) error {
	return status.Errorf(codes.Internal, "mailpit: internal server error: %s", err.Error())
}

func paramsNotFoundErr(output *oapigen.GetMessageParamsNotFoundApplicationJSON) error {
	return status.Errorf(codes.NotFound, "mailpit: message params not found: %s", *output)
}

func chaosErr(output *oapigen.GetChaosBadRequestApplicationJSON) error {
	return status.Errorf(codes.Internal, "mailpit: chaos bad request: %s", *output)
}

func unexpectedErr(output interface{}) error {
	return status.Errorf(codes.Internal, "mailpit: unexpected response: %v", output)
}

func validateParams[T any](params T) error {
	v := reflect.ValueOf(params)
	t := reflect.TypeOf(params)

	if t.Kind() != reflect.Struct {
		return status.Errorf(codes.Internal, "validateParams: expected struct, got %T", params)
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !value.CanInterface() {
			continue
		}

		if field.Type.Kind() == reflect.String {
			if value.String() == "" {
				return status.Errorf(codes.InvalidArgument, "%s is required", field.Name)
			}
		}
	}

	return nil
}
