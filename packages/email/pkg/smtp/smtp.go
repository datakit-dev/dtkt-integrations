package smtp

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/gomail.v2"

	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"

	emailintgr "github.com/datakit-dev/dtkt-integrations/email/pkg/proto/dtkt/emailintgr/v1beta1"
)

type SmtpService struct {
	config *emailintgr.Config_SmtpConfig
	dialer *gomail.Dialer
	client gomail.SendCloser
}

func NewSmtpService(config *emailintgr.Config_SmtpConfig) (*SmtpService, error) {
	dialer := gomail.NewDialer(config.Host, int(config.Port), config.Username, config.Password)
	client, err := dialer.Dial()

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to connect to SMTP server: %v", err))
	}

	return &SmtpService{
		config: config,
		dialer: dialer,
		client: client,
	}, nil
}

func (s *SmtpService) CheckConfig() error {
	client, err := s.dialer.Dial()
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to connect to SMTP server: %v", err))
	}

	err = client.Close()
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to close SMTP connection: %v", err))
	}

	return nil
}

func (s *SmtpService) Close() error {
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Failed to close SMTP client: %v", err))
		}
	}

	return nil
}

func (s *SmtpService) SendEmail(email *emailv1beta1.Email) (*emailv1beta1.EmailSendStatus, error) {
	msg, err := newEmail(email)
	if err != nil {
		return &emailv1beta1.EmailSendStatus{Error: err.Error()}, nil
	}

	err = gomail.Send(s.client, msg)
	if err == nil {
		return &emailv1beta1.EmailSendStatus{
			Success: true,
		}, nil
	}

	// If sending fails, attempt to re-dial and send again
	if s.client != nil {
		_ = s.client.Close()
	}

	newClient, err := s.dialer.Dial()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to re-dial SMTP server: %v", err))
	}

	err = gomail.Send(newClient, msg)
	if err != nil {
		return &emailv1beta1.EmailSendStatus{
			Error: fmt.Sprintf("Failed to send email: %v", err),
		}, nil
	}

	s.client = newClient

	return &emailv1beta1.EmailSendStatus{
		Success: true,
	}, nil
}
