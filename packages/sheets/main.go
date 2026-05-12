package main

import (
	"embed"
	"log"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"

	"github.com/datakit-dev/dtkt-integrations/sheets/pkg"
	sheetsv1beta "github.com/datakit-dev/dtkt-integrations/sheets/pkg/proto/integration/sheets/v1beta"
)

//go:embed package.dtkt.yaml
var spec embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(spec, pkg.NewInstance)
	if err != nil {
		log.Fatal(err)
	}

	integrationsdk.RegisterService(intgr,
		&sheetsv1beta.SpreadsheetService_ServiceDesc,
		pkg.NewSpreadsheetService,
	)

	if err := intgr.Serve(); err != nil {
		log.Fatal(err)
	}
}
