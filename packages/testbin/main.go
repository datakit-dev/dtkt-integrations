package main

import (
	"embed"
	"log"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"

	"github.com/datakit-dev/dtkt-integrations/testbin/pkg"
	testbinv1beta "github.com/datakit-dev/dtkt-integrations/testbin/pkg/proto/integration/testbin/v1beta"
)

//go:embed package.dtkt.yaml
var spec embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(spec, pkg.NewInstance)
	if err != nil {
		log.Fatal(err)
	}

	integrationsdk.RegisterService(intgr, &testbinv1beta.EchoService_ServiceDesc, pkg.NewEchoService)
	integrationsdk.RegisterService(intgr, &testbinv1beta.InspectService_ServiceDesc, pkg.NewInspectService)

	if err := intgr.Serve(); err != nil {
		log.Fatal(err)
	}
}
