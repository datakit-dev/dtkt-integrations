package v1beta1

import (
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/common"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func ParamsFromProto(params v1beta1.Params) (qparams []bigquery.QueryParameter, err error) {
	if len(params) == 0 {
		return nil, nil
	}

	qparams = make([]bigquery.QueryParameter, 0, len(params))
	for idx, p := range params {
		param, err := ParamFromProto(p)
		if err != nil {
			return nil, fmt.Errorf("params[%d] error: %w", idx, err)
		}

		qparams[idx] = *param
	}
	return
}

func ParamFromProto(p *sharedv1beta1.Param) (*bigquery.QueryParameter, error) {
	if p == nil {
		return nil, fmt.Errorf("cannot be nil")
	} else if p.Field == nil {
		return nil, fmt.Errorf("field cannot be nil")
	} else if p.Field.Name == "" {
		return nil, fmt.Errorf("field name cannot be blank")
	} else if p.Field.Type == nil {
		return nil, fmt.Errorf("field type cannot be nil")
	}

	value, err := ParamValueFromProto(p.Field, p.Value)
	if err != nil {
		return nil, err
	}

	return &bigquery.QueryParameter{
		Name:  p.Field.Name,
		Value: value,
	}, nil
}

func ArrayParamFromProto(field *sharedv1beta1.Field, anyVal *anypb.Any) (values []bigquery.QueryParameterValue, err error) {
	value, err := anyVal.UnmarshalNew()
	if err != nil {
		return nil, err
	}

	anyList, ok := value.(*structpb.ListValue)
	if !ok {
		return nil, fmt.Errorf("expected: google.protobuf.Struct, got: %T", value)
	}

	values = make([]bigquery.QueryParameterValue, len(anyList.Values))
	for idx, listVal := range anyList.Values {
		anyVal, err := common.WrapProtoAny(listVal.AsInterface())
		if err != nil {
			return nil, err
		}

		value, err := ParamValueFromProto(field, anyVal)
		if err != nil {
			return nil, err
		}

		values[idx] = *value
	}

	return
}

func StructParamFromProto(field *sharedv1beta1.Field, anyVal *anypb.Any) (typ *bigquery.StandardSQLStructType, values map[string]bigquery.QueryParameterValue, err error) {
	if len(field.Fields) == 0 {
		return nil, nil, fmt.Errorf("record field must have sub-fields")
	}

	value, err := anyVal.UnmarshalNew()
	if err != nil {
		return nil, nil, err
	}

	structVal, ok := value.(*structpb.Struct)
	if !ok {
		return nil, nil, fmt.Errorf("expected: google.protobuf.Struct, got: %T", value)
	}

	typ = &bigquery.StandardSQLStructType{
		Fields: make([]*bigquery.StandardSQLField, len(field.Fields)),
	}
	values = make(map[string]bigquery.QueryParameterValue)
	for idx, field := range field.Fields {
		if fieldVal, ok := structVal.AsMap()[field.Name]; ok {
			anyVal, err := common.WrapProtoAny(fieldVal)
			if err != nil {
				return nil, nil, err
			}

			value, err := ParamValueFromProto(field, anyVal)
			if err != nil {
				return nil, nil, err
			}
			values[field.Name] = *value

			typ.Fields[idx] = &bigquery.StandardSQLField{
				Name: field.Name,
				Type: &value.Type,
			}
		} else {
			return nil, nil, fmt.Errorf(`field "%s" must have a value`, field.Name)
		}
	}

	return
}

func ParamValueFromProto(field *sharedv1beta1.Field, anyVal *anypb.Any) (*bigquery.QueryParameterValue, error) {
	var paramVal = &bigquery.QueryParameterValue{
		Type: bigquery.StandardSQLDataType{
			TypeKind: field.Type.NativeType,
		},
	}

	if field.Repeated {
		paramVal.Type.TypeKind = "ARRAY"
		paramVal.Type.ArrayElementType = &bigquery.StandardSQLDataType{
			TypeKind: field.Type.NativeType,
		}

		arrayVal, err := ArrayParamFromProto(field, anyVal)
		if err != nil {
			return nil, err
		}

		paramVal.ArrayValue = arrayVal
		return paramVal, nil
	} else {
		switch bigquery.FieldType(field.Type.NativeType) {
		case bigquery.RecordFieldType:
			structType, structVal, err := StructParamFromProto(field, anyVal)
			if err != nil {
				return nil, err
			}

			paramVal.Type.StructType = structType
			paramVal.StructValue = structVal
			return paramVal, nil
		case bigquery.IntervalFieldType:
			stringVal, strErr := common.UnwrapProtoAnyAs[string](anyVal)
			if strErr != nil {
				duration, err := common.UnwrapProtoAnyAs[time.Duration](anyVal)
				if err != nil {
					return nil, errors.Join(strErr, err)
				}

				paramVal.Value = bigquery.IntervalValueFromDuration(duration)
				return paramVal, nil
			}

			parsed, err := bigquery.ParseInterval(stringVal)
			if err != nil {
				return nil, err
			}
			paramVal.Value = parsed

			return paramVal, nil
		case bigquery.BooleanFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullBool{}
			} else {
				anyVal, err := common.UnwrapProtoAnyAs[bool](anyVal)
				if err != nil {
					return nil, err
				}
				paramVal.Value = anyVal
			}
			return paramVal, nil
		case bigquery.StringFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullString{}
			} else {
				anyVal, err := common.UnwrapProtoAnyAs[string](anyVal)
				if err != nil {
					return nil, err
				}
				paramVal.Value = anyVal
			}
			return paramVal, nil
		case bigquery.IntegerFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullInt64{}
			} else {
				anyVal, err := common.UnwrapProtoAnyAs[int64](anyVal)
				if err != nil {
					return nil, err
				}
				paramVal.Value = anyVal
			}
			return paramVal, nil
		case bigquery.FloatFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullFloat64{}
			} else {
				anyVal, err := common.UnwrapProtoAnyAs[float64](anyVal)
				if err != nil {
					return nil, err
				}
				paramVal.Value = anyVal
			}
			return paramVal, nil
		case bigquery.GeographyFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullGeography{}
			} else {
				anyVal, err := common.UnwrapProtoAnyAs[string](anyVal)
				if err != nil {
					return nil, err
				}
				paramVal.Value = anyVal
			}
			return paramVal, nil
		case bigquery.DateFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullDate{}
				return paramVal, nil
			} else {
				anyVal, err := common.UnwrapProtoAnyAs[string](anyVal)
				if err != nil {
					return nil, err
				}
				paramVal.Value = anyVal
			}
			return paramVal, nil
		case bigquery.DateTimeFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullDateTime{}
				return paramVal, nil
			}
		case bigquery.TimeFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullTime{}
				return paramVal, nil
			}
		case bigquery.TimestampFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullTimestamp{}
				return paramVal, nil
			}
		case bigquery.JSONFieldType:
			if field.Nullable && anyVal == nil || len(anyVal.Value) == 0 {
				paramVal.Value = bigquery.NullJSON{}
				return paramVal, nil
			}
			// TODO: figure out how to handle range fields
			// case bigquery.RangeFieldType:
			// paramVal.Type.RangeElementType = &bigquery.StandardSQLDataType{
			// }
		}
	}

	return nil, fmt.Errorf("unknown field type: %s", field.Type.NativeType)
}
