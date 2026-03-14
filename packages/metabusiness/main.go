package main

import (
	"embed"

	"github.com/datakit-dev/dtkt-integrations/metabusiness/pkg"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	replicationv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/replication/v1beta1"
)

//go:embed package.dtkt.yaml
var fs embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(fs, pkg.NewInstance)
	if err != nil {
		panic(err)
	}

	integrationsdk.RegisterService(intgr, &replicationv1beta1.SourceService_ServiceDesc, pkg.NewSourceService)
	integrationsdk.RegisterManagedActionService(intgr, pkg.NewActionService, pkg.Actions()...)

	if err := intgr.Serve(); err != nil {
		panic(err)
	}
}
