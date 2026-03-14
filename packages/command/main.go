package main

import (
	"embed"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"

	"github.com/datakit-dev/dtkt-integrations/command/pkg"
)

//go:embed package.dtkt.yaml
var manifest embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(manifest, pkg.NewInstance)
	if err != nil {
		panic(err)
	}

	integrationsdk.RegisterService(intgr, &commandv1beta1.CommandService_ServiceDesc, pkg.NewCommandService)

	if err := intgr.Serve(); err != nil {
		panic(err)
	}
}
