package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"reflect"
	"strings"

	"github.com/datakit-dev/dtkt-integrations/postmark/pkg/accountoapi"
	accountoapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/accountoapi/gen"
	"github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi"
	serveroapigen "github.com/datakit-dev/dtkt-integrations/postmark/pkg/serveroapi/gen"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type inputRequestParams[Request any, Params any] struct {
	Request Request
	Params  Params
}

type inputEmpty struct{}

var internalServerErr = status.Errorf(codes.Internal, "postmark: internal server error")

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

// func parseDynamicResponse(output *serveroapigen.DynamicResponse) (any, error) {
// 	v, err := parseAny(*output)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to decode dynamic response: %v", err)
// 	}

// 	return v, nil
// }

func accountUnprocessableContentErr(output *accountoapigen.StandardPostmarkResponse) error {
	return status.Errorf(codes.InvalidArgument, "postmark: unprocessable content: %v (error code: %v)", output.Message, output.ErrorCode)
}

func serverUnprocessableContentErr(output *serveroapigen.StandardPostmarkResponse) error {
	return status.Errorf(codes.InvalidArgument, "postmark: unprocessable content: %v (error code: %v)", output.Message, output.ErrorCode)
}

func internalServerErrFromErr(err error) error {
	return status.Errorf(codes.Internal, "postmark: internal server error: %s", err.Error())
}

func unexpectedErr(output interface{}) error {
	return status.Errorf(codes.Internal, "postmark: unexpected response: %v", output)
}

func descExpr(operation string) string {
	return fmt.Sprintf(`.paths[] | .[] | select((.operationId|ascii_upcase) == ("%s"|ascii_upcase))`, operation)
}

func accountDesc(operation string) string {
	expr := descExpr(operation)
	summary, _ := accountoapi.OpenAPISpec.Value(expr + " | .summary")
	tags, _ := accountoapi.OpenAPISpec.Value(expr + ` | .tags | join(" / ")`)

	return fmt.Sprintf("%s (%s)", summary, tags)
}

func serverDesc(operation string) string {
	expr := descExpr(operation)
	summary, _ := serveroapi.OpenAPISpec.Value(expr + " | .summary")
	tags, _ := serveroapi.OpenAPISpec.Value(expr + ` | .tags | join(" / ")`)

	return fmt.Sprintf("%s (%s)", summary, tags)
}

func sendEmail(
	ctx context.Context,
	s *Instance,
	email *emailv1beta1.Email,
) (*emailv1beta1.EmailSendStatus, error) {
	res, err := s.serverClient.SendEmail(ctx, toOptSendEmailRequest(email), serveroapigen.SendEmailParams{
		XPostmarkServerToken: s.config.ServerApiToken,
	})
	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	sendStatus := &emailv1beta1.EmailSendStatus{}

	switch output := res.(type) {
	case *serveroapigen.SendEmailResponse:
		if output.ErrorCode.Value != 0 {
			sendStatus.Error = output.Message.Value
			log.Warn(ctx, "Failed to send email", slog.String("error", sendStatus.Error))

			return sendStatus, nil
		}

		log.Debug(ctx, "Sent email", slog.String("id", output.MessageID.Value))

		sendStatus.Id = output.MessageID.Value
		sendStatus.Success = true
	case *serveroapigen.StandardPostmarkResponse:
		sendStatus.Error = output.Message.Value
		log.Warn(ctx, "Failed to send email", slog.String("error", sendStatus.Error))
	case *serveroapigen.R500:
		return nil, internalServerErr
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
	res, err := s.serverClient.SendEmailWithTemplate(ctx,
		toEmailWithTemplateRequest(email),
		serveroapigen.SendEmailWithTemplateParams{
			XPostmarkServerToken: s.config.ServerApiToken,
		},
	)
	if err != nil {
		return nil, internalServerErrFromErr(err)
	}

	sendStatus := &emailv1beta1.EmailSendStatus{}

	switch output := res.(type) {
	case *serveroapigen.SendEmailResponse:
		if output.ErrorCode.Value != 0 {
			sendStatus.Error = output.Message.Value
			log.Warn(ctx, "Failed to send email with template", slog.String("error", sendStatus.Error))

			return sendStatus, nil
		}

		log.Debug(ctx, "Sent email with template", slog.String("id", output.MessageID.Value))

		sendStatus.Id = output.MessageID.Value
		sendStatus.Success = true
	case *serveroapigen.StandardPostmarkResponse:
		sendStatus.Error = output.Message.Value
		log.Warn(ctx, "Failed to send email with template", slog.String("error", sendStatus.Error))
	case *serveroapigen.R500:
		return nil, internalServerErr
	default:
		return nil, unexpectedErr(output)
	}

	return sendStatus, nil
}

func toSendEmailRequest(email *emailv1beta1.Email) serveroapigen.SendEmailRequest {
	return serveroapigen.SendEmailRequest{
		From: toEmailAddressOptString(email.From),
		To:   toEmailAddressesOptString(email.To),
		Cc:   toEmailAddressesOptString(email.Cc),
		Bcc:  toEmailAddressesOptString(email.Bcc),

		ReplyTo: toEmailAddressOptString(email.ReplyTo),

		Subject:  serveroapigen.NewOptString(email.Subject),
		HtmlBody: serveroapigen.NewOptString(email.HtmlBody),
		TextBody: serveroapigen.NewOptString(email.TextBody),

		// Tag: serveroapigen.NewOptString(""),

		TrackOpens: serveroapigen.NewOptBool(false),
		TrackLinks: serveroapigen.NewOptSendEmailRequestTrackLinks(
			serveroapigen.SendEmailRequestTrackLinksNone,
		),

		Headers:     []serveroapigen.MessageHeader{},
		Attachments: []serveroapigen.Attachment{},
	}
}

func toOptSendEmailRequest(email *emailv1beta1.Email) serveroapigen.OptSendEmailRequest {
	return serveroapigen.NewOptSendEmailRequest(toSendEmailRequest(email))
}

func toEmailWithTemplateRequest(email *emailv1beta1.EmailWithTemplate) *serveroapigen.EmailWithTemplateRequest {
	return &serveroapigen.EmailWithTemplateRequest{
		From:    toEmailAddressString(email.From),
		To:      toEmailAddressesString(email.To),
		ReplyTo: toEmailAddressOptString(email.ReplyTo),
		Cc:      toEmailAddressesOptString(email.Cc),
		Bcc:     toEmailAddressesOptString(email.Bcc),

		TemplateAlias: email.TemplateId,
		TemplateModel: email.TemplateParams,

		// InlineCss: serveroapigen.NewOptBool(true),

		Headers:     []serveroapigen.MessageHeader{},
		Attachments: []serveroapigen.Attachment{},
	}
}

func toEmailAddressString(emailAddress *emailv1beta1.EmailAddress) string {
	if emailAddress == nil {
		return ""
	}

	addr := mail.Address{Name: emailAddress.Name, Address: emailAddress.Address}
	return addr.String()
}

func toEmailAddressesString(emailAddress []*emailv1beta1.EmailAddress) string {
	var formatted []string
	for _, ea := range emailAddress {
		eas := toEmailAddressString(ea)
		if eas != "" {
			formatted = append(formatted, eas)
		}
	}

	return strings.Join(formatted, ",")
}

func toEmailAddressOptString(emailAddress *emailv1beta1.EmailAddress) serveroapigen.OptString {
	if emailAddress == nil {
		return serveroapigen.OptString{}
	}

	addr := mail.Address{Name: emailAddress.Name, Address: emailAddress.Address}
	return serveroapigen.NewOptString(addr.String())
}

func toEmailAddressesOptString(emailAddress []*emailv1beta1.EmailAddress) serveroapigen.OptString {
	var formatted []string
	for _, ea := range emailAddress {
		formatted = append(formatted, toEmailAddressString(ea))
	}

	if len(formatted) == 0 {
		return serveroapigen.OptString{}
	}
	return serveroapigen.NewOptString(strings.Join(formatted, ","))
}
