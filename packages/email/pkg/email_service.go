package pkg

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
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
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if inst.smtp == nil {
		return nil, status.Error(codes.Internal, "SMTP service is not configured")
	}

	if req.Email == nil {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	sendStatus, err := inst.smtp.SendEmail(req.Email)
	if err != nil {
		return nil, err
	}

	return &emailv1beta1.SendEmailResponse{
		SendStatus: sendStatus,
	}, nil
}

func (s *EmailService) SendEmails(stream grpc.BidiStreamingServer[emailv1beta1.SendEmailsRequest, emailv1beta1.SendEmailsResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if inst.smtp == nil {
		return status.Error(codes.Internal, "SMTP service is not configured")
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return status.Errorf(codes.Unknown, "failed to receive request: %v", err)
		}

		sendStatus, err := inst.smtp.SendEmail(req.Email)
		if err != nil {
			return err
		}

		if err := stream.Send(&emailv1beta1.SendEmailsResponse{
			SendStatus: sendStatus,
		}); err != nil {
			return status.Errorf(codes.Unknown, "failed to send response: %v", err)
		}
	}
}

func (s *EmailService) SendBatchEmail(ctx context.Context, req *emailv1beta1.SendBatchEmailRequest) (*emailv1beta1.SendBatchEmailResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if inst.smtp == nil {
		return nil, status.Error(codes.Internal, "SMTP service is not configured")
	}

	var statuses []*emailv1beta1.EmailSendStatus

	for _, email := range req.Emails {
		sendStatus, err := inst.smtp.SendEmail(email)
		if err != nil {
			return nil, err
		}

		statuses = append(statuses, sendStatus)
	}

	return &emailv1beta1.SendBatchEmailResponse{
		SendStatus: statuses,
	}, nil
}

func (s *EmailService) ListEmailTemplates(ctx context.Context, req *emailv1beta1.ListEmailTemplatesRequest) (*emailv1beta1.ListEmailTemplatesResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return inst.template.ListTemplates(req.PageSize, req.PageToken)
}

func (s *EmailService) GetEmailTemplate(ctx context.Context, req *emailv1beta1.GetEmailTemplateRequest) (*emailv1beta1.GetEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return inst.template.GetTemplate(req.Id)
}

func (s *EmailService) CreateEmailTemplate(ctx context.Context, req *emailv1beta1.CreateEmailTemplateRequest) (*emailv1beta1.CreateEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return inst.template.CreateTemplate(req.Template)
}

func (s *EmailService) UpdateEmailTemplate(ctx context.Context, req *emailv1beta1.UpdateEmailTemplateRequest) (*emailv1beta1.UpdateEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return inst.template.UpdateTemplate(req.Template)
}

func (s *EmailService) DeleteEmailTemplate(ctx context.Context, req *emailv1beta1.DeleteEmailTemplateRequest) (*emailv1beta1.DeleteEmailTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return inst.template.DeleteTemplate(req.Id)
}

func (s *EmailService) SendEmailWithTemplate(ctx context.Context, req *emailv1beta1.SendEmailWithTemplateRequest) (*emailv1beta1.SendEmailWithTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if inst.smtp == nil {
		return nil, status.Error(codes.Internal, "SMTP service is not configured")
	}

	subject, textBody, htmlBody, err := inst.template.ExecuteTemplate(req.Email.TemplateId, req.Email.TemplateParams)
	if err != nil {
		return nil, err
	}

	sendStatus, err := inst.smtp.SendEmail(&emailv1beta1.Email{
		From:       req.Email.From,
		To:         req.Email.To,
		ReplyTo:    req.Email.ReplyTo,
		Cc:         req.Email.Cc,
		Bcc:        req.Email.Bcc,
		ReturnPath: req.Email.ReturnPath,

		Subject:  subject,
		TextBody: textBody,
		HtmlBody: htmlBody,
	})
	if err != nil {
		return nil, err
	}

	return &emailv1beta1.SendEmailWithTemplateResponse{
		SendStatus: sendStatus,
	}, nil
}

func (s *EmailService) SendEmailsWithTemplate(stream grpc.BidiStreamingServer[emailv1beta1.SendEmailsWithTemplateRequest, emailv1beta1.SendEmailsWithTemplateResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if inst.smtp == nil {
		return status.Error(codes.Internal, "SMTP service is not configured")
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return status.Errorf(codes.Unknown, "failed to receive request: %v", err)
		}

		subject, textBody, htmlBody, err := inst.template.ExecuteTemplate(req.Email.TemplateId, req.Email.TemplateParams)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to execute template: %v", err)
		}

		sendStatus, err := inst.smtp.SendEmail(&emailv1beta1.Email{
			From:       req.Email.From,
			To:         req.Email.To,
			ReplyTo:    req.Email.ReplyTo,
			Cc:         req.Email.Cc,
			Bcc:        req.Email.Bcc,
			ReturnPath: req.Email.ReturnPath,

			Subject:  subject,
			TextBody: textBody,
			HtmlBody: htmlBody,
		})
		if err != nil {
			return err
		}

		if err := stream.Send(&emailv1beta1.SendEmailsWithTemplateResponse{
			SendStatus: sendStatus,
		}); err != nil {
			return status.Errorf(codes.Unknown, "failed to send response: %v", err)
		}
	}
}

func (s *EmailService) SendBatchEmailWithTemplate(ctx context.Context, req *emailv1beta1.SendBatchEmailWithTemplateRequest) (*emailv1beta1.SendBatchEmailWithTemplateResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if inst.smtp == nil {
		return nil, status.Error(codes.Internal, "SMTP service is not configured")
	}

	var statuses []*emailv1beta1.EmailSendStatus

	for _, email := range req.Emails {
		subject, textBody, htmlBody, err := inst.template.ExecuteTemplate(email.TemplateId, email.TemplateParams)
		if err != nil {
			statuses = append(statuses, &emailv1beta1.EmailSendStatus{
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		sendStatus, err := inst.smtp.SendEmail(&emailv1beta1.Email{
			From:       email.From,
			To:         email.To,
			ReplyTo:    email.ReplyTo,
			Cc:         email.Cc,
			Bcc:        email.Bcc,
			ReturnPath: email.ReturnPath,
			Subject:    subject,
			TextBody:   textBody,
			HtmlBody:   htmlBody,
		})
		if err != nil {
			return nil, err
		}

		statuses = append(statuses, sendStatus)
	}

	return &emailv1beta1.SendBatchEmailWithTemplateResponse{
		SendStatus: statuses,
	}, nil
}
