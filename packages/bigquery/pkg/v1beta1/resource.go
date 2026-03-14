package v1beta1

// import (
// 	"context"
// 	"fmt"
// 	"regexp"

// 	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
// 	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
// 	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
// 	"google.golang.org/genproto/googleapis/cloud/audit"
// )

// var (
// 	SchemaRegex = regexp.MustCompile(`projects/(?P<Project>[^/]+)/datasets/(?P<Dataset>[^/]+)`)
// 	TableRegex  = regexp.MustCompile(`projects/(?P<Project>[^/]+)/datasets/(?P<Dataset>[^/]+)/tables/(?P<Table>[^/]+)`)
// 	QueryRegex  = regexp.MustCompile("`dtkt-cloud:sqlQuery/(?P<QueryID>[^`]+)`")
// )

// func NewSchemaAction(ctx context.Context, client *lib.Client, eventType EventType, auditLog *audit.AuditLog) (sharedv1beta1.ActionType, *catalogv1beta1.Schema, error) {
// 	var (
// 		names      = ExtractMatches(SchemaRegex, auditLog.GetResourceName())
// 		actionType sharedv1beta1.ActionType
// 	)

// 	switch eventType {
// 	case CreateDatasetEvent:
// 		actionType = sharedv1beta1.ActionType_ACTION_TYPE_CREATE
// 	case UpdateDatasetEvent:
// 		actionType = sharedv1beta1.ActionType_ACTION_TYPE_UPDATE
// 	case DeleteDatasetEvent:
// 		actionType = sharedv1beta1.ActionType_ACTION_TYPE_DELETE
// 	default:
// 		return 0, nil, fmt.Errorf("unhandled schema event type: %s", eventType)
// 	}

// 	schema, err := NewDataset(client, &catalogv1beta1.Schema{
// 		Catalog: &catalogv1beta1.Catalog{
// 			Name: names["Project"],
// 		},
// 		Name: names["Dataset"],
// 	}).ToProto(ctx)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	return actionType, schema, nil
// }

// func NewTableAction(ctx context.Context, client *lib.Client, eventType EventType, auditLog *audit.AuditLog) (sharedv1beta1.ActionType, *catalogv1beta1.Table, error) {
// 	var (
// 		names      = ExtractMatches(TableRegex, auditLog.GetResourceName())
// 		actionType sharedv1beta1.ActionType
// 	)

// 	switch eventType {
// 	case CreateTableEvent:
// 		actionType = sharedv1beta1.ActionType_ACTION_TYPE_CREATE
// 	case UpdateTableEvent, PatchTableEvent:
// 		actionType = sharedv1beta1.ActionType_ACTION_TYPE_UPDATE
// 	case DeleteTableEvent:
// 		actionType = sharedv1beta1.ActionType_ACTION_TYPE_DELETE
// 	default:
// 		return 0, nil, fmt.Errorf("unhandled schema event type: %s", eventType)
// 	}

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
// }

// func ExtractMatches(re *regexp.Regexp, s string) map[string]string {
// 	matches := re.FindStringSubmatch(s)
// 	if len(matches) <= 1 {
// 		return nil
// 	}

// 	result := make(map[string]string)
// 	for i, name := range re.SubexpNames() {
// 		if i != 0 && name != "" {
// 			result[name] = matches[i]
// 		}
// 	}
// 	return result
// }
