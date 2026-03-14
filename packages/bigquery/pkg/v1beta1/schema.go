package v1beta1

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"cloud.google.com/go/bigquery"
	storage "cloud.google.com/go/bigquery/storage/apiv1"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	geov1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/geo/v1beta1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	SchemaService struct {
		catalogv1beta1.UnimplementedSchemaServiceServer
		mux v1beta1.InstanceMux[*Instance]
	}
	Dataset struct {
		dataset *bigquery.Dataset
		meta    *bigquery.DatasetMetadata
		schema  *catalogv1beta1.Schema
		tables  map[string]*Table

		mut sync.Mutex
	}
	Datasets []*Dataset
)

func NewSchemaService(mux v1beta1.InstanceMux[*Instance]) *SchemaService {
	return &SchemaService{
		mux: mux,
	}
}

func NewDataset(client *lib.Client, schema *catalogv1beta1.Schema) *Dataset {
	return NewDatasetWith(client.DatasetInProject(schema.Catalog.Name, schema.Name), schema)
}

func NewDatasetWith(dataset *bigquery.Dataset, schema *catalogv1beta1.Schema) *Dataset {
	if schema.Catalog == nil {
		schema.Catalog = &catalogv1beta1.Catalog{
			Name: dataset.ProjectID,
		}
	}

	return &Dataset{
		schema:  schema,
		dataset: dataset,
		tables:  make(map[string]*Table),
	}
}

func (s *SchemaService) ListSchemas(req *catalogv1beta1.ListSchemasRequest, stream grpc.ServerStreamingServer[catalogv1beta1.ListSchemasResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewCatalog(inst.client, req.Catalog).ListSchemas(req, stream)
}

func (s *SchemaService) GetSchema(ctx context.Context, req *catalogv1beta1.GetSchemaRequest) (*catalogv1beta1.GetSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewCatalog(inst.client, req.Catalog).GetSchema(ctx, req)
}

func (s *SchemaService) CreateSchema(ctx context.Context, req *catalogv1beta1.CreateSchemaRequest) (*catalogv1beta1.CreateSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewCatalog(inst.client, req.Catalog).CreateSchema(ctx, req)
}

func (s *SchemaService) UpdateSchema(ctx context.Context, req *catalogv1beta1.UpdateSchemaRequest) (*catalogv1beta1.UpdateSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewCatalog(inst.client, req.Catalog).UpdateSchema(ctx, req)
}

func (s *SchemaService) DeleteSchema(ctx context.Context, req *catalogv1beta1.DeleteSchemaRequest) (*catalogv1beta1.DeleteSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewCatalog(inst.client, req.Catalog).DeleteSchema(ctx, req)
}

func (d *Dataset) GetMetadata(ctx context.Context) (*bigquery.DatasetMetadata, error) {
	if d.meta == nil {
		meta, err := d.dataset.Metadata(ctx)
		if err != nil {
			return nil, err
		}
		d.meta = meta
	}
	return d.meta, nil
}

func (d *Dataset) GetLabels(ctx context.Context) (map[string]string, error) {
	metadata, err := d.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return metadata.Labels, nil
}

func (d *Dataset) ToProto(ctx context.Context) (*catalogv1beta1.Schema, error) {
	d.schema.Description = d.GetDescription(ctx)

	labels, err := d.GetLabels(ctx)
	if err != nil {
		return nil, err
	}

	v1beta1.SetCatalogLabels(d.schema.Catalog, labels)

	return d.schema, nil
}

func (d *Dataset) GetName(ctx context.Context) string {
	return d.dataset.DatasetID
}

func (d *Dataset) GetDescription(ctx context.Context) string {
	metadata, err := d.GetMetadata(ctx)
	if err != nil {
		return ""
	}
	return metadata.Description
}

func (d *Dataset) NewTable(ctx context.Context, name string, reload bool) (*Table, error) {
	d.mut.Lock()
	defer d.mut.Unlock()

	if !reload {
		if table, ok := d.tables[name]; ok {
			return table, nil
		}
	}

	tbl := d.dataset.Table(name)
	meta, err := tbl.Metadata(ctx)
	table := NewTable(tbl, meta)
	if err == nil && table.meta != nil {
		d.tables[name] = table
	}

	return table, err
}

func (d *Dataset) ListGeoSources(stream grpc.ServerStreamingServer[geov1beta1.ListGeoSourcesResponse]) error {
	var (
		iter = d.dataset.Tables(stream.Context())
	)

	for {
		tbl, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		table, err := d.NewTable(stream.Context(), tbl.TableID, false)
		if err != nil {
			return err
		}

		var (
			geoFields  []string
			propFields []*geov1beta1.PropertyField
		)
		for field := range slices.Values(table.meta.Schema) {
			switch field.Type {
			case bigquery.GeographyFieldType:
				geoFields = append(geoFields, field.Name)
			default:
				if nativeType, ok := Types.Find(string(field.Type)); ok {
					if propType, ok := v1beta1.GeoPropertyTypeFromDataType(nativeType); ok {
						propFields = append(propFields, &geov1beta1.PropertyField{
							Name: field.Name,
							Type: propType,
						})
					}
				}
			}
		}

		if len(geoFields) > 0 {
			stream.Send(&geov1beta1.ListGeoSourcesResponse{
				GeoSource: &geov1beta1.GeoSource{
					Id:         table.meta.FullID,
					GeoFields:  geoFields,
					PropFields: propFields,
					Source: &geov1beta1.GeoSource_Table{
						Table: &catalogv1beta1.Table{
							Schema: d.schema,
							Name:   table.Name(),
						},
					},
				},
			})
		}
	}

	return nil
}

func (d *Dataset) ListTables(stream grpc.ServerStreamingServer[catalogv1beta1.ListTablesResponse]) error {
	iter := d.dataset.Tables(stream.Context())

	for {
		tbl, err := iter.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}

		table, err := d.GetTable(stream.Context(), &catalogv1beta1.GetTableRequest{
			Schema: d.schema,
			Name:   tbl.TableID,
		})
		if err != nil {
			return err
		}

		err = stream.Send(&catalogv1beta1.ListTablesResponse{
			Table: table.Table,
		})
		if err != nil {
			return err
		}
	}
}

func (d *Dataset) GetTable(ctx context.Context, req *catalogv1beta1.GetTableRequest) (*catalogv1beta1.GetTableResponse, error) {
	tbl, err := d.NewTable(ctx, req.Name, true)
	if err != nil {
		return nil, err
	}

	return &catalogv1beta1.GetTableResponse{
		Table: v1beta1.NewTable(
			req.Schema,
			tbl.table.TableID,
			v1beta1.WithTableDescription(tbl.Description(ctx)),
			v1beta1.WithTableStats(&catalogv1beta1.TableStats{
				TotalRows:  int64(tbl.meta.NumRows),
				TotalBytes: int64(tbl.meta.NumBytes),
			}),
			v1beta1.WithTableFields(FieldsToProto(tbl.meta.Schema)...),
			// string(tbl.meta.Type),
		),
	}, nil
}

func (d *Dataset) CreateTable(ctx context.Context, req *catalogv1beta1.CreateTableRequest) (*catalogv1beta1.CreateTableResponse, error) {
	tbl, err := d.NewTable(ctx, req.Name, false)
	if err != nil {
		if err, ok := err.(*googleapi.Error); ok {
			if err.Code != http.StatusNotFound {
				return nil, err
			}
		}
	} else if tbl.meta != nil {
		return nil, fmt.Errorf("table already exists")
	}

	meta := &bigquery.TableMetadata{
		Description: req.Description,
	}

	if len(req.Fields) > 0 {
		meta.Schema = FieldsFromProto(req.Fields)
	}

	err = tbl.table.Create(ctx, meta)
	if err != nil {
		return nil, err
	}

	return &catalogv1beta1.CreateTableResponse{
		Table: v1beta1.NewTable(
			req.Schema,
			tbl.table.TableID,
			v1beta1.WithTableDescription(tbl.Description(ctx)),
			v1beta1.WithTableStats(&catalogv1beta1.TableStats{
				TotalRows:  int64(tbl.meta.NumRows),
				TotalBytes: int64(tbl.meta.NumBytes),
			}),
			v1beta1.WithTableFields(FieldsToProto(tbl.meta.Schema)...),
			// string(tbl.meta.Type),
		),
		Created: true,
	}, nil
}

func (d *Dataset) UpdateTable(ctx context.Context, req *catalogv1beta1.UpdateTableRequest) (*catalogv1beta1.UpdateTableResponse, error) {
	tbl, err := d.NewTable(ctx, req.Name, true)
	if err != nil {
		return nil, err
	}

	update := bigquery.TableMetadataToUpdate{
		Description: req.Description,
	}

	if len(req.Fields) > 0 {
		update.Schema = FieldsFromProto(req.Fields)
	}

	meta, err := tbl.table.Update(ctx, update, tbl.meta.ETag)
	if err != nil {
		return nil, err
	}

	return &catalogv1beta1.UpdateTableResponse{
		Updated: true,
		Table: v1beta1.NewTable(
			req.Schema,
			meta.Name,
			v1beta1.WithTableDescription(meta.Description),
			v1beta1.WithTableStats(&catalogv1beta1.TableStats{
				TotalRows:  int64(meta.NumRows),
				TotalBytes: int64(meta.NumBytes),
			}),
			v1beta1.WithTableFields(FieldsToProto(meta.Schema)...),
			// string(tbl.meta.Type),
		),
	}, nil
}

func (d *Dataset) DeleteTable(ctx context.Context, req *catalogv1beta1.DeleteTableRequest) (*catalogv1beta1.DeleteTableResponse, error) {
	d.mut.Lock()
	defer d.mut.Unlock()

	tbl, err := d.NewTable(ctx, req.Name, true)
	if err != nil {
		return nil, err
	}

	var msg string
	err = tbl.table.Delete(ctx)
	if err == nil {
		delete(d.tables, req.Name)
	} else {
		msg = err.Error()
	}

	return &catalogv1beta1.DeleteTableResponse{
		Table: &catalogv1beta1.Table{
			Schema: req.Schema,
			Name:   tbl.table.TableID,
		},
		Deleted: err == nil,
		Error:   msg,
	}, nil
}

func (d *Dataset) ReadTable(req *catalogv1beta1.ReadTableRequest, stream grpc.ServerStreamingServer[catalogv1beta1.ReadTableResponse], reader *storage.BigQueryReadClient, opts *storagepb.ReadSession_TableReadOptions) error {
	table, err := d.NewTable(stream.Context(), req.Name, false)
	if err != nil {
		return err
	}

	var readRow = func(result resultRow) error {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			rowMap, err := result.Row()
			if err != nil {
				return err
			}

			err = stream.Send(&catalogv1beta1.ReadTableResponse{
				Table: &catalogv1beta1.Table{
					Schema: req.Schema,
					Name:   req.Name,
				},
				Row: rowMap,
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	return table.Read(stream.Context(), reader, readRow, opts)
}
