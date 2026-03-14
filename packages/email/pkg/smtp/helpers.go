package smtp

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/gomail.v2"

	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
)

func formatAddresses(msg *gomail.Message, addresses []*emailv1beta1.EmailAddress) []string {
	to := make([]string, len(addresses))

	for i, addr := range addresses {
		to[i] = msg.FormatAddress(addr.Address, addr.Name)
	}

	return to
}

func newEmail(email *emailv1beta1.Email) (*gomail.Message, error) {
	msg := gomail.NewMessage()

	if email.From == nil {
		return nil, status.Error(codes.InvalidArgument, "Email.From is required")
	}

	msg.SetAddressHeader("From", email.From.Address, email.From.Name)

	if email.ReturnPath != nil {
		msg.SetAddressHeader("Return-Path", email.ReturnPath.Address, email.ReturnPath.Name)
	}

	if email.To != nil {
		msg.SetHeader("To", formatAddresses(msg, email.To)...)
	}

	if email.Cc != nil {
		msg.SetHeader("Cc", formatAddresses(msg, email.Cc)...)
	}

	if email.Bcc != nil {
		msg.SetHeader("Bcc", formatAddresses(msg, email.Bcc)...)
	}

	if email.Subject == "" {
		return nil, status.Error(codes.InvalidArgument, "Email.Subject is required")
	}

	msg.SetHeader("Subject", email.Subject)

	if email.TextBody == "" {
		return nil, status.Error(codes.InvalidArgument, "Email.TextBody is required")
	}

	msg.SetBody("text/plain", email.TextBody)

	if email.HtmlBody != "" {
		msg.AddAlternative("text/html", email.HtmlBody)
	}

	return msg, nil
}
