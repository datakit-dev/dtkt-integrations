package v1beta1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	CatalogService struct {
		catalogv1beta1.UnimplementedCatalogServiceServer
		mux v1beta1.InstanceMux[*Instance]
	}
	Catalog struct {
		client  *lib.Client
		catalog *catalogv1beta1.Catalog
	}
)

func NewCatalogService(mux v1beta1.InstanceMux[*Instance]) *CatalogService {
	return &CatalogService{
		mux: mux,
	}
}

func NewCatalog(client *lib.Client, catalog *catalogv1beta1.Catalog) *Catalog {
	return &Catalog{
		client:  client,
		catalog: catalog,
	}
}

func (s *CatalogService) ListDataTypes(context.Context, *catalogv1beta1.ListDataTypesRequest) (*catalogv1beta1.ListDataTypesResponse, error) {
	return v1beta1.NewDataTypesResponse(Types), nil
}

func (s *CatalogService) GetCatalog(ctx context.Context, req *catalogv1beta1.GetCatalogRequest) (*catalogv1beta1.GetCatalogResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if req == nil {
		req = &catalogv1beta1.GetCatalogRequest{}
	}

	if req.Name == "" {
		req.Name = inst.client.Config.ProjectId
	}

	ok, err := NewCatalog(inst.client, &catalogv1beta1.Catalog{
		Name:     req.Name,
		Metadata: req.Metadata,
	}).Check(ctx)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("catalog not found: %s", req.Name)
	}

	return &catalogv1beta1.GetCatalogResponse{
		Catalog: v1beta1.NewCatalog(
			req.Name,
			v1beta1.WithCatalogMetadata(req.Metadata),
		),
	}, nil
}

func (c *Catalog) GetDatasets(ctx context.Context) (*bigquery.DatasetIterator, error) {
	it := c.client.Datasets(ctx)
	it.ProjectID = c.catalog.Name
	if labels := v1beta1.GetCatalogLabels(c.catalog); len(labels) > 0 {
		var labelStr []string
		for key, value := range labels {
			labelStr = append(labelStr, fmt.Sprintf("labels.%s:%s", key, value))
		}
		it.Filter = strings.Join(labelStr, " ")
	}
	return it, nil
}

func (c *Catalog) GetDataset(req *catalogv1beta1.GetSchemaRequest) *Dataset {
	return NewDataset(c.client, &catalogv1beta1.Schema{
		Catalog: req.Catalog,
		Name:    req.Name,
	})
}

func (c *Catalog) Check(ctx context.Context) (bool, error) {
	it, err := c.GetDatasets(ctx)
	if err != nil {
		return false, err
	}

	_, err = it.Next()
	if err != nil && err != iterator.Done {
		return false, err
	}

	return true, nil
}

func (c *Catalog) ListSchemas(req *catalogv1beta1.ListSchemasRequest, stream grpc.ServerStreamingServer[catalogv1beta1.ListSchemasResponse]) error {
	iter, err := c.GetDatasets(stream.Context())
	if err != nil {
		return err
	}

	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			return nil
		} else if err != nil {
			return err
		}

		schema, err := c.GetSchema(stream.Context(), &catalogv1beta1.GetSchemaRequest{
			Catalog: req.Catalog,
			Name:    ds.DatasetID,
		})
		if err != nil {
			return err
		}

		err = stream.Send(&catalogv1beta1.ListSchemasResponse{
			Schema: schema.Schema,
		})
		if err != nil {
			return err
		}
	}
}

func (c *Catalog) GetSchema(ctx context.Context, req *catalogv1beta1.GetSchemaRequest) (*catalogv1beta1.GetSchemaResponse, error) {
	var ds = c.GetDataset(req)
	_, err := ds.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	if givenLabels := v1beta1.GetCatalogLabels(c.catalog); len(givenLabels) > 0 {
		appliedLabels, err := ds.GetLabels(ctx)
		if err != nil {
			return nil, err
		}

		for givenKey, givenVal := range givenLabels {
			for appliedKey, appliedVal := range appliedLabels {
				if givenKey == appliedKey && givenVal != appliedVal {
					return nil, fmt.Errorf("schema not found")
				}
			}
		}
	}

	return &catalogv1beta1.GetSchemaResponse{
		Schema: v1beta1.NewSchema(req.Catalog, ds.GetName(ctx), v1beta1.WithSchemaDescription(ds.GetDescription(ctx))),
	}, nil
}

func (c *Catalog) CreateSchema(ctx context.Context, req *catalogv1beta1.CreateSchemaRequest) (*catalogv1beta1.CreateSchemaResponse, error) {
	var ds = c.GetDataset(&catalogv1beta1.GetSchemaRequest{
		Catalog: c.catalog,
		Name:    req.Name,
	})

	meta, err := ds.GetMetadata(ctx)
	if err != nil {
		if err, ok := err.(*googleapi.Error); ok {
			if err.Code != http.StatusNotFound {
				return nil, err
			}
		}
	} else if meta != nil {
		return nil, fmt.Errorf("schema already exists")
	}

	meta = &bigquery.DatasetMetadata{
		Description: req.Description,
	}

	if labels := v1beta1.GetCatalogLabels(c.catalog); len(labels) > 0 {
		meta.Labels = labels
	}

	if err := ds.dataset.Create(ctx, meta); err != nil {
		return nil, err
	}

	return &catalogv1beta1.CreateSchemaResponse{
		Created: true,
		Schema:  v1beta1.NewSchema(req.Catalog, ds.dataset.DatasetID, v1beta1.WithSchemaDescription(meta.Description)),
	}, nil
}

func (c *Catalog) UpdateSchema(ctx context.Context, req *catalogv1beta1.UpdateSchemaRequest) (*catalogv1beta1.UpdateSchemaResponse, error) {
	ds := c.GetDataset(&catalogv1beta1.GetSchemaRequest{
		Catalog: c.catalog,
		Name:    req.Name,
	})

	read, err := ds.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	update, err := ds.dataset.Update(ctx, bigquery.DatasetMetadataToUpdate{
		Description: req.Description,
	}, read.ETag)
	if err != nil {
		return nil, err
	}

	return &catalogv1beta1.UpdateSchemaResponse{
		Updated: true,
		Schema:  v1beta1.NewSchema(req.Catalog, ds.dataset.DatasetID, v1beta1.WithSchemaDescription(update.Description)),
	}, nil
}

func (c *Catalog) DeleteSchema(ctx context.Context, req *catalogv1beta1.DeleteSchemaRequest) (*catalogv1beta1.DeleteSchemaResponse, error) {
	ds := c.GetDataset(&catalogv1beta1.GetSchemaRequest{
		Catalog: c.catalog,
		Name:    req.Name,
	})

	var (
		err = ds.dataset.DeleteWithContents(ctx)
		msg string
	)

	if err != nil {
		msg = err.Error()
	}

	return &catalogv1beta1.DeleteSchemaResponse{
		Schema:  v1beta1.NewSchema(req.Catalog, ds.dataset.DatasetID),
		Deleted: err == nil,
		Error:   msg,
	}, nil
}
