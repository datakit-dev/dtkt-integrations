package pkg

import (
	"context"

	emailintgr "github.com/datakit-dev/dtkt-integrations/email/pkg/proto/dtkt/emailintgr/v1beta1"
	"github.com/datakit-dev/dtkt-integrations/email/pkg/smtp"
	"github.com/datakit-dev/dtkt-integrations/email/pkg/template"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Integration instance struct
type Instance struct {
	config   *emailintgr.Config
	smtp     *smtp.SmtpService
	template *template.TemplateService
}

// Creates a new instance
func NewInstance(ctx context.Context, config *emailintgr.Config) (*Instance, error) {
	var err error
	var (
		smtpService *smtp.SmtpService
	)

	if config.Smtp != nil {
		smtpService, err = smtp.NewSmtpService(config.Smtp)
		if err != nil {
			return nil, err
		}

		err = smtpService.CheckConfig()
		if err != nil {
			return nil, err
		}
	}

	template, err := template.NewTemplateService(config.Templates)
	if err != nil {
		return nil, err
	}

	return &Instance{
		config:   config,
		smtp:     smtpService,
		template: template,
	}, nil
}

func (i *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

// Close is called to close instance resources (e.g. client connections)
func (i *Instance) Close() error {
	if i.smtp != nil {
		if err := i.smtp.Close(); err != nil {
			return err
		}
	}
	return nil
}
