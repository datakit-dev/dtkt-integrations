package main

import (
	"embed"
	"log"
	"os"

	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/env"
	"google.golang.org/grpc"

	"github.com/datakit-dev/dtkt-integrations/browser/pkg"
)

//go:embed package.dtkt.yaml
var spec embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(spec, pkg.NewInstance)
	if err != nil {
		log.Fatal(err)
	}

	extArg, ok := pkg.HasExtensionArg(os.Args...)
	if ok {
		err = pkg.StartExtension(extArg)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if env.GetVar(env.AppEnv) == "dev" {
			err = pkg.InstallExtension(intgr.Logger())
			if err != nil {
				log.Fatal(err)
			}
		}

		taskSvc, err := pkg.NewTaskService(intgr)
		if err != nil {
			log.Fatal(err)
		}

		schemaSvc, err := pkg.NewExtractionSchemaService(intgr)
		if err != nil {
			log.Fatal(err)
		}

		recordSvc, err := pkg.NewExtractionRecordService(intgr)
		if err != nil {
			log.Fatal(err)
		}

		intgr.RegisterService(&browserv1beta.TaskService_ServiceDesc, func(reg grpc.ServiceRegistrar) {
			browserv1beta.RegisterTaskServiceServer(reg, taskSvc)
		})

		intgr.RegisterService(&browserv1beta.ExtractionSchemaService_ServiceDesc, func(reg grpc.ServiceRegistrar) {
			browserv1beta.RegisterExtractionSchemaServiceServer(reg, schemaSvc)
		})

		intgr.RegisterService(&browserv1beta.ExtractionRecordService_ServiceDesc, func(reg grpc.ServiceRegistrar) {
			browserv1beta.RegisterExtractionRecordServiceServer(reg, recordSvc)
		})

		intgr.RegisterProxy(&browserv1beta.ChromeService_ServiceDesc, pkg.GetExtensionClient(intgr))

		if err := intgr.Serve(); err != nil {
			log.Fatal(err)
		}
	}
}
