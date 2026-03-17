package main

import (
	"embed"
	"log"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"

	"github.com/datakit-dev/dtkt-integrations/fivetran/pkg"

	replicationv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/replication/v1beta1"
)

//go:embed package.dtkt.yaml
var spec embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(spec, pkg.NewInstance)
	if err != nil {
		log.Fatal(err)
	}

	integrationsdk.RegisterService(intgr, &replicationv1beta1.DestinationService_ServiceDesc, pkg.NewDestinationService)
	integrationsdk.RegisterService(intgr, &replicationv1beta1.SourceService_ServiceDesc, pkg.NewSourceService)
	integrationsdk.RegisterManagedActionService(intgr, pkg.NewActionService, pkg.Actions...)
	integrationsdk.RegisterManagedEventService(intgr, pkg.NewEventService, pkg.Events, pkg.EventSources...)

	if err := intgr.Serve(); err != nil {
		log.Fatal(err)
	}
}
