package zetasql

import (
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
	zetatypes "github.com/goccy/go-zetasql/types"
)

var ZetaTypes = map[bigquery.FieldType]zetatypes.Type{
	bigquery.BooleanFieldType:    zetatypes.BoolType(),
	bigquery.StringFieldType:     zetatypes.StringType(),
	bigquery.BytesFieldType:      zetatypes.BytesType(),
	bigquery.IntegerFieldType:    zetatypes.Int64Type(),
	bigquery.FloatFieldType:      zetatypes.DoubleType(),
	bigquery.NumericFieldType:    zetatypes.NumericType(),
	bigquery.BigNumericFieldType: zetatypes.BigNumericType(),
	bigquery.DateFieldType:       zetatypes.DateType(),
	bigquery.DateTimeFieldType:   zetatypes.DatetimeType(),
	bigquery.TimeFieldType:       zetatypes.TimeType(),
	bigquery.TimestampFieldType:  zetatypes.TimestampType(),
	bigquery.JSONFieldType:       zetatypes.JsonType(),
	bigquery.RecordFieldType:     zetatypes.EmptyStructType(),
	bigquery.GeographyFieldType:  zetatypes.GeographyType(),
	bigquery.IntervalFieldType:   zetatypes.IntervalType(),
	// Range type is not supported by ZetaSQL and experimental in BigQuery
	// string(bigquery.RangeFieldType),
}

func NewField(name string, zetaType zetatypes.Type) (field *sharedv1beta1.Field, err error) {
	var (
		typeKind   = zetaType.Kind()
		isRepeated = zetaType.IsArray()
	)

	if zetaType.IsArray() {
		typeKind = zetaType.AsArray().ElementType().Kind()
	}

	var found bool
	for bqType, zType := range ZetaTypes {
		found = zType.Kind() == typeKind
		if found {
			field = v1beta1.NewField(name,
				v1beta1.NewDataType(string(bqType)),
				v1beta1.WithFieldRepeated(isRepeated),
			)

			if zetaType.HasAnyFields() {
				for _, f := range zetaType.AsStruct().Fields() {
					sf, err := NewField(f.Name(), f.Type())
					if err != nil {
						return field, err
					}
					field.Fields = append(field.Fields, sf)
				}
			}
			return
		}
	}
	return field, fmt.Errorf("zetasql.NewField unhandled field %s with type: %s", name, zetaType.DebugString(true))
}
