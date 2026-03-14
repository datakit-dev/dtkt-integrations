package pkg

import (
	"context"
	"log/slog"

	oapigen "github.com/datakit-dev/dtkt-integrations/mailpit/pkg/oapi/gen"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
)

func sendEmail(
	ctx context.Context,
	s *Instance,
	email *emailv1beta1.Email,
) (*emailv1beta1.EmailSendStatus, error) {
	res, err := s.client.SendMessageParams(ctx, toOptSendMessageParamsReq(email))
	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	sendStatus := &emailv1beta1.EmailSendStatus{}

	switch output := res.(type) {
	case *oapigen.SendMessageResponse:
		log.Debug(ctx, "Sent email", slog.String("id", output.ID.Value))
		sendStatus.Success = true
	case *oapigen.JSONErrorResponse:
		sendStatus.Error = output.Error.Value
		log.Error(ctx, "Failed to send email", slog.String("error", sendStatus.Error))
	default:
		return nil, unexpectedErr(output)
	}

	return sendStatus, nil
}

func sendEmailWithTemplate(
	ctx context.Context,
	s *Instance,
	email *emailv1beta1.EmailWithTemplate,
) (*emailv1beta1.EmailSendStatus, error) {
	subject, textBody, htmlBody, err := s.template.ExecuteTemplate(email.TemplateId, email.TemplateParams)
	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	res, err := s.client.SendMessageParams(ctx,
		toOptSendMessageParamsReq(
			templateEmailToEmail(email, subject, textBody, htmlBody),
		),
	)

	sendStatus := &emailv1beta1.EmailSendStatus{}

	if err != nil {
		sendStatus.Error = err.Error()
	}

	switch output := res.(type) {
	case *oapigen.SendMessageResponse:
		log.Debug(ctx, "Sent email with template", slog.String("id", output.ID.Value))
		sendStatus.Success = true
	case *oapigen.JSONErrorResponse:
		sendStatus.Error = output.Error.Value
	default:
		return nil, unexpectedErr(output)
	}

	return sendStatus, nil
}

func templateEmailToEmail(
	email *emailv1beta1.EmailWithTemplate,
	subject string,
	textBody string,
	htmlBody string,
) *emailv1beta1.Email {
	return &emailv1beta1.Email{
		From:       email.From,
		To:         email.To,
		ReplyTo:    email.ReplyTo,
		Cc:         email.Cc,
		Bcc:        email.Bcc,
		ReturnPath: email.ReturnPath,

		Subject:  subject,
		TextBody: textBody,
		HtmlBody: htmlBody,
	}
}

func toOptSendMessageParamsReq(email *emailv1beta1.Email) oapigen.OptSendMessageParamsReq {
	bcc := make([]string, 0, len(email.Bcc))
	for _, b := range email.Bcc {
		bcc = append(bcc, b.Address)
	}

	cc := make([]oapigen.SendMessageParamsReqCcItem, 0, len(email.Cc))
	for _, c := range email.Cc {
		cc = append(cc, oapigen.SendMessageParamsReqCcItem{
			Email: c.Address,
			Name:  oapigen.NewOptString(c.Name),
		})
	}

	to := make([]oapigen.SendMessageParamsReqToItem, 0, len(email.To))
	for _, t := range email.To {
		to = append(to, oapigen.SendMessageParamsReqToItem{
			Email: t.Address,
			Name:  oapigen.NewOptString(t.Name),
		})
	}

	from := oapigen.SendMessageParamsReqFrom{
		Email: email.From.Address,
		Name:  oapigen.NewOptString(email.From.Name),
	}

	replyTo := make([]oapigen.SendMessageParamsReqReplyToItem, 0)
	if email.ReplyTo != nil {
		replyTo = append(replyTo, oapigen.SendMessageParamsReqReplyToItem{
			Email: email.ReplyTo.Address,
			Name:  oapigen.NewOptString(email.ReplyTo.Name),
		})
	}

	return oapigen.NewOptSendMessageParamsReq(oapigen.SendMessageParamsReq{
		Attachments: []oapigen.SendMessageParamsReqAttachmentsItem{},

		Bcc:     bcc,
		Cc:      cc,
		From:    from,
		To:      to,
		ReplyTo: replyTo,
		Tags:    []string{},
		Subject: oapigen.NewOptString(email.Subject),
		HTML:    oapigen.NewOptString(email.HtmlBody),
		Text:    oapigen.NewOptString(email.TextBody),
	})
}
