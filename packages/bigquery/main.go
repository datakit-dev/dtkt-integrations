package main

import (
	"embed"
	"log"

	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	pkgv1beta1 "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/v1beta1"
	fivetran "github.com/datakit-dev/dtkt-integrations/fivetran/lib"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	geov1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/geo/v1beta1"
	replicationv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/replication/v1beta1"
)

//go:embed package.dtkt.yaml
var pkg embed.FS

func main() {
	intgr, err := integrationsdk.NewFS(pkg, pkgv1beta1.NewInstance)
	if err != nil {
		log.Fatal(err)
	}

	integrationsdk.RegisterService(intgr, &catalogv1beta1.CatalogService_ServiceDesc, pkgv1beta1.NewCatalogService)
	integrationsdk.RegisterService(intgr, &catalogv1beta1.SchemaService_ServiceDesc, pkgv1beta1.NewSchemaService)
	integrationsdk.RegisterService(intgr, &catalogv1beta1.TableService_ServiceDesc, pkgv1beta1.NewTableService)
	integrationsdk.RegisterService(intgr, &catalogv1beta1.QueryService_ServiceDesc, pkgv1beta1.NewQueryService)
	integrationsdk.RegisterService(intgr, &geov1beta1.GeoService_ServiceDesc, pkgv1beta1.NewGeoService)

	destService := fivetran.NewDestinationService(intgr,
		&replicationv1beta1.DestinationType{
			Id:          "big_query",
			Name:        intgr.Package().GetIdentity().GetName(),
			Description: intgr.Package().GetDescription(),
			IconUrl:     intgr.Package().GetIcon(),
			Category:    "Warehouse",
		},
	)

	events := lib.AuditLogEvents[*pkgv1beta1.Instance]()
	events = append(events,
		fivetran.WebhookEvents[*pkgv1beta1.Instance]()...,
	)

	integrationsdk.RegisterService(intgr, &replicationv1beta1.DestinationService_ServiceDesc,
		func(mux v1beta1.InstanceMux[*pkgv1beta1.Instance]) replicationv1beta1.DestinationServiceServer {
			return destService
		},
	)

	integrationsdk.RegisterManagedEventService(intgr, pkgv1beta1.NewEventService, events,
		lib.AuditLogSource[*pkgv1beta1.Instance](),
		fivetran.WebhookEventSource(intgr),
	)

	if err := intgr.Serve(); err != nil {
		log.Fatal(err)
	}
}
