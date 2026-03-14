package main

import (
	"embed"

	"github.com/datakit-dev/dtkt-integrations/localblob/pkg"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	blobv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/blob/v1beta1"
)

//go:embed package.dtkt.yaml
var manifest embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(manifest, pkg.NewInstance)
	if err != nil {
		panic(err)
	}

	integrationsdk.RegisterService(intgr, &blobv1beta1.BlobService_ServiceDesc, pkg.NewBlobService)

	if err := intgr.Serve(); err != nil {
		panic(err)
	}
}
