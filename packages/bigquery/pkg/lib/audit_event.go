package lib

import (
	"context"
	"regexp"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"cloud.google.com/go/pubsub"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// Dataset (Schema) events:
	CreateDatasetEvent EventType = "google.cloud.bigquery.v2.DatasetService.InsertDataset"
	UpdateDatasetEvent EventType = "google.cloud.bigquery.v2.DatasetService.UpdateDataset"
	PatchDatasetEvent  EventType = "google.cloud.bigquery.v2.DatasetService.PatchDataset"
	DeleteDatasetEvent EventType = "google.cloud.bigquery.v2.DatasetService.DeleteDataset"

	// Table events:
	CreateTableEvent EventType = "google.cloud.bigquery.v2.TableService.InsertTable"
	UpdateTableEvent EventType = "google.cloud.bigquery.v2.TableService.UpdateTable"
	PatchTableEvent  EventType = "google.cloud.bigquery.v2.TableService.PatchTable"
	DeleteTableEvent EventType = "google.cloud.bigquery.v2.TableService.DeleteTable"
	ReadRowsEvent    EventType = "google.cloud.bigquery.storage.v1.BigQueryRead.ReadRows"
	AppendRowsEvent  EventType = "google.cloud.bigquery.storage.v1.BigQueryWrite.AppendRows"

	// TBD handling these events:
	// QueryEvent EventType = "google.cloud.bigquery.v2.JobService.Query"
	// ReadTableEvent         EventType = "google.cloud.bigquery.v2.TableDataService.List"
	// InsertJobEvent EventType = "google.cloud.bigquery.v2.JobService.InsertJob"
	// QueryResultsEvent      EventType = "google.cloud.bigquery.v2.JobService.GetQueryResults"
	// CreateReadSessionEvent EventType = "google.cloud.bigquery.storage.v1.BigQueryRead.CreateReadSession"
	// SplitReadStreamEvent   EventType = "google.cloud.bigquery.storage.v1.BigQueryRead.SplitReadStream"
)

var (
	SchemaRegex = regexp.MustCompile(`projects/(?P<projectId>[^/]+)/datasets/(?P<datasetId>[^/]+)`)
	TableRegex  = regexp.MustCompile(`projects/(?P<projectId>[^/]+)/datasets/(?P<datasetId>[^/]+)/tables/(?P<tableId>[^/]+)`)
	QueryRegex  = regexp.MustCompile("`dtkt-cloud:sqlQuery/(?P<queryId>[^`]+)`")
)

type (
	EventType  string
	EventNames struct {
		ProjectID string
		DatasetID string
		TableID   string
		QueryID   string
	}
)

func AuditLogEvents[I v1beta1.InstanceType]() []v1beta1.RegisterEventFunc[I] {
	return []v1beta1.RegisterEventFunc[I]{
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Schema](
			CreateDatasetEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_CREATE,
			"Event triggered when a dataset is created",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Schema](
			UpdateDatasetEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_UPDATE,
			"Event triggered when a dataset is updated",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Schema](
			PatchDatasetEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_UPDATE,
			"Event triggered when a dataset is patched",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Schema](
			DeleteDatasetEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_DELETE,
			"Event triggered when a dataset is deleted",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Table](
			CreateTableEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_CREATE,
			"Event triggered when a table is created",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Table](
			UpdateTableEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_UPDATE,
			"Event triggered when a table is updated",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Table](
			PatchTableEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_UPDATE,
			"Event triggered when a table is patched",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Table](
			DeleteTableEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_DELETE,
			"Event triggered when a table is deleted",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Table](
			ReadRowsEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_READ,
			"Event triggered when data is read from a table",
		),
		v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Table](
			AppendRowsEvent,
			sharedv1beta1.ActionType_ACTION_TYPE_WRITE,
			"Event triggered when data is written to a table",
		),
		// v1beta1.RegisterEventWithAction[I, *catalogv1beta1.Query](
		// 	QueryEvent,
		// 	sharedv1beta1.ActionType_ACTION_TYPE_CREATE,
		// 	"Event triggered when a query job is created",
		// ),
		// v1beta1.RegisterEvent[I, *catalogv1beta1.Query](
		// 	InsertJobEvent,
		// 	"Event triggered when a job is created",
		// ),
	}
}

func ProcessAuditLogEvent[I AuditLogInstance](ctx context.Context, inst I, events *v1beta1.EventRegistry, msg *pubsub.Message) (*v1beta1.EventWithPayload, error) {
	logEntry := &loggingpb.LogEntry{}
	err := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(msg.Data, logEntry)
	if err != nil {
		return nil, err
	}

	log := &audit.AuditLog{}
	if err := logEntry.GetProtoPayload().UnmarshalTo(log); err != nil {
		return nil, err
	} else if log.GetMetadata() == nil {
		return nil, err
	}

	event, err := events.Find(log.GetMethodName())
	if err != nil {
		return nil, err
	}

	payload, err := inst.GetAuditLogEvent(ctx, event, log)
	if err != nil {
		return nil, err
	}

	return event.WithPayload(payload)
}

func ExtractNames(re *regexp.Regexp, s string) EventNames {
	matches := ExtractMatches(re, s)
	return EventNames{
		ProjectID: matches["projectId"],
		DatasetID: matches["datasetId"],
		TableID:   matches["tableId"],
		QueryID:   matches["queryId"],
	}
}

func ExtractMatches(re *regexp.Regexp, s string) map[string]string {
	matches := re.FindStringSubmatch(s)
	if len(matches) <= 1 {
		return nil
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}
	return result
}

func (m EventType) String() string {
	return string(m)
}
