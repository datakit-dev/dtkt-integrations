package v1beta1

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1/geojsonwrapperpb"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	geov1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/geo/v1beta1"
	geojsonv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/geojson/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
	"github.com/twpayne/go-geom/encoding/wkt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GeoService struct {
	geov1beta1.UnimplementedGeoServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewGeoService(mux v1beta1.InstanceMux[*Instance]) *GeoService {
	return &GeoService{
		mux: mux,
	}
}

func (s *GeoService) ListGeoSources(req *geov1beta1.ListGeoSourcesRequest, stream grpc.ServerStreamingServer[geov1beta1.ListGeoSourcesResponse]) error {
	_, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	// var defaultCatalog *catalogv1beta1.Catalog
	// if req.GetCatalog().GetName() == "" || req.GetSchema().GetCatalog().GetName() == "" {
	// 	// resp, err := i.GetCatalog(stream.Context(), &catalogv1beta1.GetCatalogRequest{})
	// 	// if err != nil {
	// 	// 	return err
	// 	// }
	// 	// defaultCatalog = resp.Catalog
	// }

	// if req.GetSchema() != nil {
	// 	if req.GetSchema().Name == "" {
	// 		return fmt.Errorf("schema name is required")
	// 	}

	// 	if req.GetSchema().GetCatalog() == nil || req.GetSchema().GetCatalog().Name == "" {
	// 		if defaultCatalog == nil {
	// 			return fmt.Errorf("catalog is required")
	// 		}

	// 		req.GetSchema().Catalog = defaultCatalog
	// 	}

	// 	return NewDataset(inst.client, req.GetSchema()).ListGeoSources(stream)
	// } else {
	// 	if req.GetCatalog() != nil && req.GetCatalog().Name != "" {
	// 		defaultCatalog = req.GetCatalog()
	// 	} else if defaultCatalog == nil {
	// 		return fmt.Errorf("catalog is required")
	// 	}

	// 	iter, err := NewCatalog(inst.client, defaultCatalog).GetDatasets(stream.Context())
	// 	if err != nil {
	// 		return err
	// 	}

	// 	for {
	// 		dataset, err := iter.Next()
	// 		if err == iterator.Done {
	// 			break
	// 		}
	// 		if err != nil {
	// 			return err
	// 		}

	// 		err = NewDatasetWith(dataset, &catalogv1beta1.Schema{
	// 			Catalog: defaultCatalog,
	// 			Name:    dataset.DatasetID,
	// 		}).ListGeoSources(stream)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}

func (s *GeoService) StreamGeoJson(req *geov1beta1.StreamGeoJsonRequest, stream grpc.ServerStreamingServer[geov1beta1.StreamGeoJsonResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if req.GetSource() != nil && req.GetSource().GetTable() != nil {
		if req.GetGeoField() == "" {
			return fmt.Errorf("source table geo field is required")
		}

		table, err := NewDataset(inst.client, req.GetSource().GetTable().Schema).
			NewTable(stream.Context(), req.GetSource().GetTable().Name, true)
		if err != nil {
			return err
		}

		if tableType := table.Type(stream.Context()); tableType != nil && *tableType != string(bigquery.RegularTable) {
			ident, err := table.Identifier(bigquery.StandardSQLID)
			if err != nil {
				return err
			}

			var propFields = []string{req.GeoField}
			for field := range slices.Values(req.GetSource().PropFields) {
				propFields = append(propFields, field.Name)
			}

			req.Source.Source = &geov1beta1.GeoSource_Query{
				Query: &catalogv1beta1.Query{
					Dialect: QueryDialect,
					Query: fmt.Sprintf("SELECT %s FROM %s",
						strings.Join(propFields, ", "), ident,
					),
				},
			}
		} else {
			var rowCount int64 = 1
			var readRow = func(result resultRow) error {
				defer func() {
					rowCount += 1
				}()

				if geoWKT, ok := result[req.GeoField].(string); ok {
					geom, err := wkt.Unmarshal(geoWKT)
					if err != nil {
						return err
					}

					delete(result, req.GeoField)
					feature, err := geojsonwrapperpb.NewFeatureFromGeom(rowCount, geom, result).ToProto()
					if err != nil {
						return err
					}

					return stream.Send(&geov1beta1.StreamGeoJsonResponse{
						Result: &geojsonv1beta1.GeoJSON{
							Geojson: &geojsonv1beta1.GeoJSON_Feature{
								Feature: feature,
							},
						},
					})
				} else {
					return fmt.Errorf("expected valid GeoJSON object, got: %T", result[req.GeoField])
				}
			}

			return readTableGeoJSON(stream.Context(), inst, table, req.GetSource(), req.GeoField, req.PropFields, req.Bounds, readRow)
		}
	}

	if req.GetSource() != nil && req.GetSource().GetQuery() != nil {
		var query = req.GetSource().GetQuery()
		if query == nil || query.GetQuery() == "" {
			return fmt.Errorf("source query is required")
		} else if req.GetGeoField() == "" {
			return fmt.Errorf("source query geo field is required")
		}

		var readRow = func(row resultRow) error {
			if geoJSON, ok := row["geoJSON"].(string); ok {
				result, err := geojsonwrapperpb.UnmarshalGeoJSON("Feature", []byte(geoJSON))
				if err != nil {
					return err
				}

				return stream.Send(&geov1beta1.StreamGeoJsonResponse{
					Result: result,
				})
			} else {
				return fmt.Errorf("expected valid GeoJSON object, got: %T", row["geoJSON"])
			}
		}

		return readQueryGeoJSON(stream.Context(), inst.client, query, req.GetSource(), req.GeoField, req.PropFields, req.Bounds, readRow)
	}

	return fmt.Errorf("unsupported GeoJSON source: %s", req.Source)
}

func getBoundsGeoJSON(bounds *geov1beta1.Bounds) *geojsonwrapperpb.GeoJSON {
	return geojsonwrapperpb.NewGeoJSON(&geojsonv1beta1.GeoJSON{
		Geojson: &geojsonv1beta1.GeoJSON_Geometry{
			Geometry: bounds.Geom,
		},
	})
}

func getBoundsExpr(bounds *geov1beta1.Bounds, geoField, geoJSON string) (string, error) {
	if bounds.GetType() == geov1beta1.BoundsType_BOUNDS_TYPE_UNSPECIFIED {
		return "", fmt.Errorf("bounds type is required")
	} else if bounds.GetGeom() == nil {
		return "", fmt.Errorf("bounds geometry is required")
	}

	var boundsFunc string
	switch bounds.Type {
	case geov1beta1.BoundsType_BOUNDS_TYPE_COVERS:
		boundsFunc = "ST_COVERS"
	case geov1beta1.BoundsType_BOUNDS_TYPE_INTERSECTS:
		boundsFunc = "ST_INTERSECTS"
	case geov1beta1.BoundsType_BOUNDS_TYPE_WITHIN:
		boundsFunc = "ST_WITHIN"
	default:
		return "", fmt.Errorf("unsupported bounds type: %s", bounds.Type)
	}

	if bounds.Centroid {
		return fmt.Sprintf("%s(ST_GEOGFROMGEOJSON(@%s), ST_CENTROID(%s))", boundsFunc, geoJSON, geoField), nil
	}
	return fmt.Sprintf("%s(ST_GEOGFROMGEOJSON(@%s), %s)", boundsFunc, geoJSON, geoField), nil
}

func validateGeoFields(source *geov1beta1.GeoSource, geoField string, propFields []string, schema bigquery.Schema) (geoColumn *sharedv1beta1.Field, propColumns v1beta1.Fields, err error) {
	var propsMissing []string
	for _, field := range FieldsToProto(schema) {
		if strings.EqualFold(field.Name, geoField) && field.Type.NativeType == string(bigquery.GeographyFieldType) {
			geoColumn = field
		}

		if len(propFields) > 0 {
			var columnFound = slices.ContainsFunc(propFields, func(name string) bool {
				return strings.EqualFold(field.Name, name)
			})

			if !columnFound {
				propsMissing = append(propsMissing, field.Name)
				continue
			}

			dataType, ok := Types.Find(field.Type.NativeType)
			if !ok {
				return nil, nil, fmt.Errorf("data type not found: %s", field.Type)
			}

			_, ok = v1beta1.GeoPropertyTypeFromDataType(dataType)
			if !ok {
				return nil, nil, fmt.Errorf("invalid field type: %s", field.Type)
			}

			propColumns = append(propColumns, field)
		}
	}

	if geoColumn == nil {
		return nil, nil, fmt.Errorf("missing geo field: %s", geoField)
	} else if len(propColumns) != len(source.GetPropFields()) {
		return nil, nil, fmt.Errorf("missing prop fields(s): %s", strings.Join(propsMissing, ", "))
	}

	return
}

func readTableGeoJSON(ctx context.Context, inst *Instance, table *Table, source *geov1beta1.GeoSource, geoField string, propFields []string, bounds *geov1beta1.Bounds, readRow readRowFunc) error {
	geoColumn, propColumns, err := validateGeoFields(source, geoField, propFields, table.meta.Schema)
	if err != nil {
		return err
	}

	var opts = &storagepb.ReadSession_TableReadOptions{
		SelectedFields: []string{geoColumn.Name},
	}
	if len(propColumns) > 0 {
		for col := range slices.Values(propColumns) {
			opts.SelectedFields = append(opts.SelectedFields, col.Name)
		}
	}

	if bounds != nil {
		geoJSON, err := getBoundsGeoJSON(bounds).MarshalJSON()
		if err != nil {
			return err
		}

		boundsExpr, err := getBoundsExpr(bounds, geoColumn.Name, string(geoJSON))
		if err != nil {
			return err
		}

		opts.RowRestriction = boundsExpr
	}

	reader, err := NewTableReader(ctx, inst.client)
	if err != nil {
		return err
	}

	return table.Read(ctx, reader, readRow, opts)
}

func readQueryGeoJSON(ctx context.Context, client *lib.Client, query *catalogv1beta1.Query, source *geov1beta1.GeoSource, geoField string, propFields []string, bounds *geov1beta1.Bounds, readRow readRowFunc) error {
	var queryBuf strings.Builder
	if bounds != nil {
		geoJSON, err := getBoundsGeoJSON(bounds).MarshalJSON()
		if err != nil {
			return err
		}

		var geomParam = "geom"
		boundsExpr, err := getBoundsExpr(bounds, geoField, geomParam)
		if err != nil {
			return err
		}

		if param, err := v1beta1.NewParamWithValue(
			v1beta1.NewField(geomParam,
				v1beta1.NewDataType(bigquery.StringFieldType),
			), string(geoJSON)); err != nil {
			return err
		} else {
			query.Params = append(query.Params, param)
		}

		queryBuf.WriteString(fmt.Sprintf(`
			WITH 	preBoundsCte AS (%s),
						boundsCte AS (SELECT *, %s AS isBounded FROM preBoundsCte),
						baseCte AS (SELECT * EXCEPT(isBounded) FROM boundsCte WHERE isBounded),
		`, query.GetQuery(), boundsExpr))
	} else {
		queryBuf.WriteString(fmt.Sprintf("WITH baseCte AS (%s),", query.GetQuery()))
	}

	if len(source.GetPropFields()) == 0 {
		queryBuf.WriteString(fmt.Sprintf(`
			dataCte AS (
				SELECT PARSE_JSON(ST_ASGEOJSON(%s), wide_number_mode => 'round') AS geom
				FROM baseCte
			)
			SELECT JSON_SET(JSON '{"type": "Feature"}', '$.geometry', geom) AS geoJSON FROM dataCte
		`, geoField))
	} else {
		queryBuf.WriteString(fmt.Sprintf(`
			dataCte AS (
				SELECT PARSE_JSON(ST_ASGEOJSON(%s), wide_number_mode => 'round') AS geom, TO_JSON(STRUCT(%s)) AS props
				FROM baseCte
			)
			SELECT JSON_SET(JSON '{"type": "Feature"}', '$.geometry', geom, '$.properties', props) AS geoJSON FROM dataCte
		`, geoField, strings.Join(propFields, ", ")))
	}

	var q = NewQuery(client, queryBuf.String(), query.Params)
	if err := q.runValidation(ctx); err != nil {
		return err
	} else if _, _, err = validateGeoFields(source, geoField, propFields, q.schema); err != nil {
		return err
	}

	return q.StreamResults(ctx, readRow)
}
