package main

import (
	"embed"

	"github.com/datakit-dev/dtkt-integrations/openai/pkg"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	aiv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/ai/v1beta1"
)

//go:embed package.dtkt.yaml
var fs embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(fs, pkg.NewInstance)
	if err != nil {
		panic(err)
	}

	integrationsdk.RegisterService(intgr, &aiv1beta1.EmbeddingService_ServiceDesc, pkg.NewEmbeddingService)
	integrationsdk.RegisterManagedActionService(intgr, pkg.NewActionService, pkg.RealtimeActions()...)
	integrationsdk.RegisterManagedEventService(intgr, pkg.NewEventService, pkg.RealtimeEvents(), pkg.RealtimeSource())

	if err := intgr.Serve(); err != nil {
		panic(err)
	}
}
