package v1beta1_test

// import (
// 	"context"
// 	"testing"

// 	bigqueryv1beta "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/proto/integration/bigquery/v1beta"
// 	pkgv1beta1 "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/v1beta1"
// 	"github.com/datakit-dev/dtkt-integrations/bigquery/test"
// 	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
// 	"github.com/datakit-dev/dtkt-sdk/sdk-go/middleware"
// 	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
// 	"github.com/goccy/bigquery-emulator/server"
// 	"google.golang.org/api/option"
// )

// var (
// 	ctx          context.Context
// 	projectID    = "datakit_test"
// 	testServer   *server.TestServer
// 	testInstance *pkgv1beta1.Instance
// )

// func init() {
// 	em, err := server.New(server.MemoryStorage)
// 	if err != nil {
// 		panic(err)
// 	}

// 	if err = em.Load(server.StructSource(test.BigQueryProjects...)); err != nil {
// 		panic(err)
// 	}

// 	if err = em.SetProject(projectID); err != nil {
// 		panic(err)
// 	}

// 	//nolint:errcheck // Impossible to return an err
// 	em.SetLogLevel(server.LogLevelFatal)

// 	//nolint:errcheck // Impossible to return an err
// 	em.SetLogFormat(server.LogFormatConsole)

// 	testServer = em.TestServer()

// 	ctx = middleware.NewRequestContext(
// 		context.Background(),
// 		middleware.NewRequest(
// 			"",
// 			"",
// 			0,
// 		),
// 	)

// 	testInstance, err = pkgv1beta1.NewInstanceWithOptions(ctx, &bigqueryv1beta.Config{
// 		ProjectId: projectID,
// 	},
// 		option.WithEndpoint(testServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithTelemetryDisabled(),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func TestService_GetCatalog(t *testing.T) {
// 	if _, err := testInstance.GetCatalog(ctx, &catalogv1beta1.GetCatalogRequest{
// 		Name: "datakit-test",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_DataTypes(t *testing.T) {
// 	if _, err := testInstance.ListDataTypes(ctx, nil); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_QueryDialect(t *testing.T) {
// 	if _, err := testInstance.GetQueryDialect(ctx, nil); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_QueryValidate(t *testing.T) {
// 	if _, err := testInstance.ValidateQuery(ctx, &catalogv1beta1.ValidateQueryRequest{
// 		Query: "SELECT * FROM `datakit-test.test_schema.test_table`",
// 		Accessible: []*catalogv1beta1.CatalogPermission{
// 			{
// 				Name: "datakit-test",
// 				Schemas: []*catalogv1beta1.SchemaPermission{
// 					{
// 						Name: "test_schema",
// 						Tables: []*catalogv1beta1.TablePermission{
// 							{
// 								Name: "test_table",
// 								Columns: []*catalogv1beta1.ColumnPermission{
// 									{
// 										Name: "id",
// 										Type: "INTEGER",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_QueryResults(t *testing.T) {
// 	if res, err := testInstance.GetQueryResults(ctx, &catalogv1beta1.GetQueryResultsRequest{
// 		Query: "SELECT * FROM `datakit-test.test_schema.test_table`",
// 	}); err != nil {
// 		t.Error(err)
// 	} else {
// 		t.Log(res)
// 	}
// }

// func TestService_GetSchemas(t *testing.T) {
// 	// ctx, cancel := context.WithCancel(ctx)
// 	// defer cancel()

// 	// _, dialer := grpcmock.MockServerWithBufConn(
// 	// 	grpcmock.RegisterService(func(s grpc.ServiceRegistrar, srv catalogv1beta1.CatalogServiceServer) {
// 	// 		s.RegisterService(&catalogv1beta1.CatalogService_ServiceDesc, testService)
// 	// 	}),
// 	// 	func(s *grpcmock.Server) {
// 	// 		s.ExpectServerStream("catalog.v1beta1.CatalogService/GetSchemas")
// 	// 	},
// 	// )(t)

// 	// var actual = make([]*catalogv1beta1.GetCatalogResponse, 0)

// 	// stream, err := testClient.GetSchemas(ctx, &catalogv1beta1.GetSchemasRequest{
// 	// 	Catalog: &catalogv1beta1.Catalog{
// 	// 		Name: "datakit-test",
// 	// 	},
// 	// })
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	// for {
// 	// 	schema, err := stream.Recv()
// 	// 	if err != nil {
// 	// 		t.Fatal(err)
// 	// 	}
// 	// 	t.Log(schema)
// 	// }

// 	// err := grpcmock.InvokeServerStream(ctx, "catalog.v1beta.CatalogService/GetSchemas", &catalogv1beta1.GetSchemasRequest{
// 	// 	Catalog: &catalogv1beta1.Catalog{
// 	// 		Name: "datakit-test",
// 	// 	},
// 	// },
// 	// 	func(s grpc.ClientStream) error {

// 	// 		fmt.Println("ClientStream:", s)
// 	// 		return nil
// 	// 	},
// 	// )
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// testService.GetSchemas(&catalogv1beta1.GetSchemasRequest{
// 	// 	Catalog: &catalogv1beta1.Catalog{
// 	// 		Name: "datakit-test",
// 	// 	},
// 	// }, stream)
// 	// if err := ; err != nil {
// 	// 	t.Error(err)
// 	// } else {
// 	// 	for schema := range schCh {
// 	// 		t.Log(schema)
// 	// 	}
// 	// }
// }

// func TestService_GetSchema(t *testing.T) {
// 	if _, err := testInstance.GetSchema(ctx, &catalogv1beta1.GetSchemaRequest{
// 		Catalog: &catalogv1beta1.Catalog{
// 			Name: "datakit-test",
// 		},
// 		Name: "test",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_CreateSchema(t *testing.T) {
// 	if _, err := testInstance.CreateSchema(ctx, &catalogv1beta1.CreateSchemaRequest{
// 		Catalog: &catalogv1beta1.Catalog{
// 			Name: "datakit-test",
// 		},
// 		Name:        "test_schema2",
// 		Description: "test",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_UpdateSchema(t *testing.T) {
// 	if _, err := testInstance.UpdateSchema(ctx, &catalogv1beta1.UpdateSchemaRequest{
// 		Catalog: &catalogv1beta1.Catalog{
// 			Name: "datakit-test",
// 		},
// 		Name:        "test_schema",
// 		Description: "test!",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_DeleteSchema(t *testing.T) {
// 	if _, err := testInstance.DeleteSchema(ctx, &catalogv1beta1.DeleteSchemaRequest{
// 		Catalog: &catalogv1beta1.Catalog{
// 			Name: "datakit-test",
// 		},
// 		Name: "test_schema2",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// // func TestService_GetTables(t *testing.T) {
// // 	ctx, cancel := context.WithCancel(ctx)
// // 	defer cancel()

// // 	if tableCh, err := testService.GetTables(ctx, &catalogv1beta1.GetTablesRequest{
// // 		Schema: &catalogv1beta1.GetSchemaRequest{
// // 			Catalog: &catalogv1beta1.Catalog{
// // 				Name: "datakit-test",
// // 			},
// // 			Name: "test_schema",
// // 		},
// // 	}); err != nil {
// // 		t.Error(err)
// // 	} else {
// // 		for table := range tableCh {
// // 			t.Log(table)
// // 		}
// // 	}
// // }

// func TestService_GetTable(t *testing.T) {
// 	if _, err := testInstance.GetTable(ctx, &catalogv1beta1.GetTableRequest{
// 		Schema: &catalogv1beta1.Schema{
// 			Catalog: &catalogv1beta1.Catalog{
// 				Name: "datakit-test",
// 			},
// 			Name: "test_schema",
// 		},
// 		Name: "test_table",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_CreateTable(t *testing.T) {
// 	if _, err := testInstance.CreateTable(ctx, &catalogv1beta1.CreateTableRequest{
// 		Schema: &catalogv1beta1.Schema{
// 			Catalog: &catalogv1beta1.Catalog{
// 				Name: "datakit-test",
// 			},
// 			Name: "test_schema",
// 		},
// 		Name:        "test_table2",
// 		Description: "test",
// 		Fields: v1beta1.Fields{
// 			v1beta1.NewField("id", v1beta1.NewDataType("INTEGER")),
// 		},
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_UpdateTable(t *testing.T) {
// 	if _, err := testInstance.UpdateTable(ctx, &catalogv1beta1.UpdateTableRequest{
// 		Schema: &catalogv1beta1.Schema{
// 			Catalog: &catalogv1beta1.Catalog{
// 				Name: "datakit-test",
// 			},
// 			Name: "test_schema",
// 		},
// 		Name:        "test_table",
// 		Description: "test!",
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestService_DeleteTable(t *testing.T) {
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	if resp, err := testInstance.DeleteTable(ctx, &catalogv1beta1.DeleteTableRequest{
// 		Schema: &catalogv1beta1.Schema{
// 			Catalog: &catalogv1beta1.Catalog{
// 				Name: "datakit-test",
// 			},
// 			Name: "test_schema",
// 		},
// 		Name: "deletable_table",
// 	}); err != nil {
// 		t.Error(err)
// 	} else if !resp.Deleted {
// 		t.Error("DeleteTable failed")
// 	}
// }
