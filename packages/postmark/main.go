package main

import (
	"embed"

	"github.com/datakit-dev/dtkt-integrations/postmark/pkg"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	emailv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/email/v1beta1"
)

//go:embed package.dtkt.yaml
var fs embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(fs, pkg.NewInstance)
	if err != nil {
		panic(err)
	}

	integrationsdk.RegisterService(intgr, &emailv1beta1.EmailService_ServiceDesc, pkg.NewEmailService)
	integrationsdk.RegisterManagedActionService(intgr, pkg.NewActionService, pkg.Actions()...)

	if err := intgr.Serve(); err != nil {
		panic(err)
	}
}
