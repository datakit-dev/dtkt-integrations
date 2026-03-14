package v1beta1

import (
	"context"
	"fmt"

	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewAuditLogEvent(ctx context.Context, auditEvent v1beta1.RegisteredEvent, auditLog *audit.AuditLog) (any, error) {
	eventType := lib.EventType(auditEvent.Proto().GetDisplayName())
	switch eventType {
	case lib.CreateDatasetEvent, lib.UpdateDatasetEvent, lib.DeleteDatasetEvent:
		names := lib.ExtractNames(lib.SchemaRegex, auditLog.GetResourceName())
		// names := lib.ExtractNames(lib.SchemaRegex, auditLog.GetResourceName())

		//	schema, err := NewDataset(client, &catalogv1beta1.Schema{
		//		Catalog: &catalogv1beta1.Catalog{
		//			Name: names["Project"],
		//		},
		//		Name: names["Dataset"],
		//	}).ToProto(ctx)
		//
		//	if err != nil {
		//		return 0, nil, err
		//	}

		return &catalogv1beta1.Schema{
			Name: fmt.Sprintf("catalogs/%s/schemas/%s", names.ProjectID, names.DatasetID),
			// SchemaId: names.DatasetID,
		}, nil
	case lib.CreateTableEvent, lib.UpdateTableEvent, lib.PatchTableEvent, lib.DeleteTableEvent:
		names := lib.ExtractNames(lib.TableRegex, auditLog.GetResourceName())

		// 	schema, err := NewDataset(client, &catalogv1beta1.Schema{
		// 		Catalog: &catalogv1beta1.Catalog{
		// 			Name: names["Project"],
		// 		},
		// 		Name: names["Dataset"],
		// 	}).ToProto(ctx)
		// 	if err != nil {
		// 		return 0, nil, err
		// 	}

		// 	return actionType, &catalogv1beta1.Table{
		// 		Schema: schema,
		// 		Name:   names["Table"],
		// 	}, nil

		return &catalogv1beta1.Table{
			Name: fmt.Sprintf("catalogs/%s/schemas/%s/tables/%s", names.ProjectID, names.DatasetID, names.TableID),
			// TableId: names.TableID,
		}, nil
		// case lib.ReadRowsEvent:
		// case lib.AppendRowsEvent:
		// case lib.QueryEvent:
		// case bigquery.InsertJobEvent:
		// 	// TBD what we do with job insertion (can be many types of jobs)
		// case bigquery.QueryResultsEvent:
		// 	// TBD what we do with query result
		// case bigquery.QueryEvent:
		// 	return UpdateSQLQuery(ctx, env, metadataJson, node.OnActionEvent(
		// 		node.WithEventType(property.SQLQueryRun),
		// 		node.WithEventMetadata(metadataEvent.GetEvent()),
		// 	))
		// case bigquery.CreateDatasetEvent:
		// 	return SyncSchema(ctx, env, resourceName, node.OnActionEvent(
		// 		node.WithEventMetadata(metadataEvent.GetDatasetChange()),
		// 	))
		// case bigquery.UpdateDatasetEvent, bigquery.PatchDatasetEvent:
		// 	return SyncSchema(ctx, env, resourceName, node.OnActionEvent(
		// 		node.WithEventMetadata(metadataEvent.GetDatasetChange()),
		// 	))
		// case bigquery.DeleteDatasetEvent:
		// 	schema, err := GetSchema(ctx, env.Ent, resourceName)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	return SendSignal(ctx, env, schema, property.SchemaDeleted,
		// 		node.WithEventMetadata(metadataEvent.GetDatasetDeletion()),
		// 	)
		// case bigquery.CreateTableEvent:
		// 	// Sync table if it exists, otherwise sync schema
		// 	err := SyncTable(ctx, env, resourceName, node.OnActionEvent(
		// 		node.WithEventMetadata(metadataEvent.GetTableCreation()),
		// 	))
		// 	if err != nil {
		// 		if ent.IsNotFound(err) {
		// 			return SyncSchema(ctx, env, resourceName, node.OnActionEvent(
		// 				node.WithEventMetadata(metadataEvent.GetTableCreation()),
		// 			))
		// 		}
		// 	}
		// 	return err
		// case bigquery.UpdateTableEvent, bigquery.PatchTableEvent:
		// 	return SyncTable(ctx, env, resourceName, node.OnActionEvent(
		// 		node.WithEventMetadata(metadataEvent.GetTableChange()),
		// 	))
		// case bigquery.DeleteTableEvent:
		// 	table, err := GetTable(ctx, env.Ent, resourceName)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	return SendSignal(ctx, env, table, property.TableDeleted,
		// 		node.WithEventMetadata(metadataEvent.GetTableDeletion()),
		// 	)
		// 	// Events below use the AuditLog format
		// 	//  - resource.type=bigquery_table BigQuery Storage API
		// 	//  - Edge case: bigquery.CreateReadSessionEvent (is bigquery_dataset) and bigquery.SplitReadStreamEvent (we can ignore these events)
		// case bigquery.ReadRowsEvent:
		// 	if logEntry.Operation != nil && logEntry.Operation.Last {
		// 		table, err := GetTable(ctx, env.Ent, resourceName)
		// 		if err != nil {
		// 			return err
		// 		}

		// 		if metadata, ok := property.JSONValue[map[string]any](metadataJson, ".tableDataRead | { fields, rowCount }"); ok {
		// 			return PublishEvent(ctx, env, table, property.TableRead,
		// 				node.WithEventMetadata(metadata),
		// 			)
		// 		}
		// 	}
		// case bigquery.AppendRowsEvent:
		// 	if logEntry.Operation != nil && logEntry.Operation.Last {
		// 		if metadata, ok := property.JSONValue[map[string]any](metadataJson, ".tableDataChange | { insertedRowsCount }"); ok {
		// 			return SyncTable(ctx, env, resourceName,
		// 				node.OnActionEvent(
		// 					node.WithEventType(property.TableWrite),
		// 					node.WithEventMetadata(metadata),
		// 				),
		// 			)
		// 		}
		// 	}
	}
	return nil, status.Errorf(codes.Unimplemented, "unhandled audit log event: %q", eventType)
}
