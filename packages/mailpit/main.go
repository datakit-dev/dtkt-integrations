package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/datakit-dev/dtkt-integrations/mailpit/pkg"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
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

	mailpitEnv := os.Environ()
	mailpitEnv = append(mailpitEnv,
		"MP_ENABLE_CHAOS=true",
		"MP_POP3_AUTH=user:password",
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mailpitCmd := exec.CommandContext(ctx, "mailpit")
	mailpitCmd.Env = mailpitEnv
	mailpitCmd.Stdout = os.Stdout
	mailpitCmd.Stderr = os.Stderr

	if err := mailpitCmd.Start(); err != nil {
		slog.Error("Failed to start Mailpit", log.Err(err))
		panic(err)
	}

	if err := intgr.Serve(); err != nil {
		panic(err)
	}
}
