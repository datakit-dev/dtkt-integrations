package pkg

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strconv"

	serveroapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi/gen"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmailService struct {
	emailv1beta1.UnimplementedEmailServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewEmailService(mux v1beta1.InstanceMux[*Instance]) *EmailService {
	return &EmailService{
		mux: mux,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, req *emailv1beta1.SendEmailRequest) (*emailv1beta1.SendEmailResponse, error) {
	log.Debug(ctx, "Sending email", slog.Any("email", req.Email))

	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	sendStatus, err := sendEmail(ctx, inst, req.Email)
	if err != nil {
		return nil, err
	}

	return &emailv1beta1.SendEmailResponse{
		SendStatus: sendStatus,
	}, nil
}

func (s *EmailService) SendEmails(stream grpc.BidiStreamingServer[emailv1beta1.SendEmailsRequest, emailv1beta1.SendEmailsResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return status.Errorf(codes.Unknown, "failed to receive request: %v", err)
		}

		sendStatus, err := sendEmail(ctx, inst, req.Email)
		if err != nil {
			return err
		}

		if err := stream.Send(&emailv1beta1.SendEmailsResponse{
			SendStatus: sendStatus,
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}
}

func (s *EmailService) SendBatchEmail(ctx context.Context, req *emailv1beta1.SendBatchEmailRequest) (*emailv1beta1.SendBatchEmailResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	var statuses []*emailv1beta1.EmailSendStatus
	for _, email := range req.Emails {
		sendStatus, err := sendEmail(ctx, inst, email)
		if err != nil {
			return nil, err
		}

		statuses = append(statuses, sendStatus)
	}

	emails := make([]serveroapigen.SendEmailRequest, 0, len(req.Emails))

	for _, email := range req.Emails {
		emails = append(emails, toSendEmailRequest(email))
	}

	res, err := inst.serverClient.SendEmailBatch(ctx, emails, serveroapigen.SendEmailBatchParams{
		XPostmarkServerToken: inst.config.ServerApiToken,
	})
	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.SendEmailBatchResponse:
		for _, emailResponse := range *output {
			sendStatus := &emailv1beta1.EmailSendStatus{}
			if emailResponse.ErrorCode.Value != 0 {
				sendStatus.Error = emailResponse.Message.Value
			} else {
				sendStatus.Id = emailResponse.MessageID.Value
				sendStatus.Success = true
			}
			statuses = append(statuses, sendStatus)
		}

		return &emailv1beta1.SendBatchEmailResponse{
			SendStatus: statuses,
		}, nil
	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}
}

func (s *EmailService) ListEmailTemplates(ctx context.Context, req *emailv1beta1.ListEmailTemplatesRequest) (*emailv1beta1.ListEmailTemplatesResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	offset := 0
	if req.PageToken != "" {
		o, err := strconv.Atoi(req.PageToken)
		if err != nil {
			log.Error(ctx, "Invalid page token", log.Err(err))
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token: %s", err.Error())
		}
		offset = o
	}

	count := min(int(req.PageSize), 1000)
	if count == 0 {
		count = 50
	}

	res, err := inst.serverClient.ListTemplates(ctx, serveroapigen.ListTemplatesParams{
		Count:  count,
		Offset: offset,

		XPostmarkServerToken: inst.config.ServerApiToken,
	})

	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.TemplateListingResponse:
		templates := make([]*emailv1beta1.EmailTemplate, 0, len(output.Templates))
		for _, template := range output.Templates {
			res, err := inst.serverClient.GetSingleTemplate(ctx, serveroapigen.GetSingleTemplateParams{
				TemplateIdOrAlias: template.Alias.Value,

				XPostmarkServerToken: inst.config.ServerApiToken,
			})

			if err != nil {
				return nil, internalServerErrFromErr(err)
			}

			switch output := res.(type) {
			case *serveroapigen.TemplateDetailResponse:
				templates = append(templates, &emailv1beta1.EmailTemplate{
					Id:       output.Alias.Value,
					Language: emailv1beta1.TemplateLanguage_TEMPLATE_LANGUAGE_MUSTACHIO,
					Subject:  output.Subject.Value,
					TextBody: output.TextBody.Value,
					HtmlBody: output.HtmlBody.Value,
				})
			case *serveroapigen.StandardPostmarkResponse:
				return nil, serverUnprocessableContentErr(output)
			case *serveroapigen.R500:
				return nil, internalServerErr
			default:
				return nil, unexpectedErr(output)
			}
		}

		nextPageToken := ""
		if len(templates) == count {
			nextPageToken = strconv.Itoa(offset + count)
		}

		return &emailv1beta1.ListEmailTemplatesResponse{
			Templates:     templates,
			NextPageToken: nextPageToken,
		}, nil
	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}
}

func (s *EmailService) GetEmailTemplate(ctx context.Context, req *emailv1beta1.GetEmailTemplateRequest) (*emailv1beta1.GetEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	res, err := inst.serverClient.GetSingleTemplate(ctx, serveroapigen.GetSingleTemplateParams{
		TemplateIdOrAlias: req.Id,

		XPostmarkServerToken: inst.config.ServerApiToken,
	})

	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.TemplateDetailResponse:
		return &emailv1beta1.GetEmailTemplateResponse{
			Template: &emailv1beta1.EmailTemplate{
				Id:       output.Alias.Value,
				Language: emailv1beta1.TemplateLanguage_TEMPLATE_LANGUAGE_MUSTACHIO,
				Subject:  output.Subject.Value,
				TextBody: output.TextBody.Value,
				HtmlBody: output.HtmlBody.Value,
			},
		}, nil
	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}
}

func (s *EmailService) CreateEmailTemplate(ctx context.Context, req *emailv1beta1.CreateEmailTemplateRequest) (*emailv1beta1.CreateEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if req.Template.Language != emailv1beta1.TemplateLanguage_TEMPLATE_LANGUAGE_MUSTACHIO {
		return nil, status.Errorf(codes.InvalidArgument, "[postmark] unsupported template language: %s", req.Template.Language.String())
	}

	res, err := inst.serverClient.CreateTemplate(ctx, &serveroapigen.CreateTemplateRequest{
		Alias:    serveroapigen.NewOptString(req.Template.Id),
		Name:     req.Template.Id,
		Subject:  req.Template.Subject,
		HtmlBody: serveroapigen.NewOptString(req.Template.HtmlBody),
		TextBody: serveroapigen.NewOptString(req.Template.TextBody),
	}, serveroapigen.CreateTemplateParams{
		XPostmarkServerToken: inst.config.ServerApiToken,
	})

	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.TemplateRecordResponse:
		return &emailv1beta1.CreateEmailTemplateResponse{
			Template: req.Template,
		}, nil

	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}
}

func (s *EmailService) UpdateEmailTemplate(ctx context.Context, req *emailv1beta1.UpdateEmailTemplateRequest) (*emailv1beta1.UpdateEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	res, err := inst.serverClient.UpdateTemplate(ctx, &serveroapigen.EditTemplateRequest{
		Alias:    serveroapigen.NewOptString(req.Id),
		Subject:  serveroapigen.NewOptString(req.Template.Subject),
		HtmlBody: serveroapigen.NewOptString(req.Template.HtmlBody),
		TextBody: serveroapigen.NewOptString(req.Template.TextBody),
	}, serveroapigen.UpdateTemplateParams{
		TemplateIdOrAlias: req.Id,

		XPostmarkServerToken: inst.config.ServerApiToken,
	})

	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.TemplateRecordResponse:
		return &emailv1beta1.UpdateEmailTemplateResponse{
			Template: &emailv1beta1.EmailTemplate{
				Id:       output.Alias.Value,
				Language: emailv1beta1.TemplateLanguage_TEMPLATE_LANGUAGE_MUSTACHIO,
				Subject:  req.Template.Subject,
				TextBody: req.Template.TextBody,
				HtmlBody: req.Template.HtmlBody,
			},
		}, nil
	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}
}

func (s *EmailService) DeleteEmailTemplate(ctx context.Context, req *emailv1beta1.DeleteEmailTemplateRequest) (*emailv1beta1.DeleteEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	res, err := inst.serverClient.DeleteTemplate(ctx, serveroapigen.DeleteTemplateParams{
		TemplateIdOrAlias: req.Id,

		XPostmarkServerToken: inst.config.ServerApiToken,
	})
	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.TemplateDetailResponse:
		return &emailv1beta1.DeleteEmailTemplateResponse{
			Id: output.Alias.Value,
		}, nil
	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}

}

func (s *EmailService) SendEmailWithTemplate(ctx context.Context, req *emailv1beta1.SendEmailWithTemplateRequest) (*emailv1beta1.SendEmailWithTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	sendStatus, err := sendEmailWithTemplate(ctx, inst, req.Email)
	if err != nil {
		return nil, err
	}

	return &emailv1beta1.SendEmailWithTemplateResponse{
		SendStatus: sendStatus,
	}, nil
}

func (s *EmailService) SendEmailsWithTemplate(stream grpc.BidiStreamingServer[emailv1beta1.SendEmailsWithTemplateRequest, emailv1beta1.SendEmailsWithTemplateResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive request: %v", err)
		}

		sendStatus, err := sendEmailWithTemplate(ctx, inst, req.Email)
		if err != nil {
			return err
		}

		if err := stream.Send(&emailv1beta1.SendEmailsWithTemplateResponse{
			SendStatus: sendStatus,
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}
}

func (s *EmailService) SendBatchEmailWithTemplate(ctx context.Context, req *emailv1beta1.SendBatchEmailWithTemplateRequest) (*emailv1beta1.SendBatchEmailWithTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	var statuses []*emailv1beta1.EmailSendStatus
	emails := make([]serveroapigen.EmailWithTemplateRequest, 0, len(req.Emails))
	for _, email := range req.Emails {
		emails = append(emails, *toEmailWithTemplateRequest(email))
	}

	res, err := inst.serverClient.SendEmailBatchWithTemplates(ctx, &serveroapigen.SendEmailTemplatedBatchRequest{
		Messages: emails,
	}, serveroapigen.SendEmailBatchWithTemplatesParams{
		XPostmarkServerToken: inst.config.ServerApiToken,
	})

	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	switch output := res.(type) {
	case *serveroapigen.SendEmailBatchResponse:
		for _, emailResponse := range *output {
			sendStatus := &emailv1beta1.EmailSendStatus{}
			if emailResponse.ErrorCode.Value != 0 {
				sendStatus.Error = emailResponse.Message.Value
			} else {
				sendStatus.Id = emailResponse.MessageID.Value
				sendStatus.Success = true
			}
			statuses = append(statuses, sendStatus)
		}

		return &emailv1beta1.SendBatchEmailWithTemplateResponse{
			SendStatus: statuses,
		}, nil
	case *serveroapigen.StandardPostmarkResponse:
		return nil, serverUnprocessableContentErr(output)
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}
}
