package v1beta1

import (
	"slices"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	sharedv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/shared/v1beta1"
)

func FieldsToProto(schema bigquery.Schema) (fields v1beta1.Fields) {
	for subField := range slices.Values(schema) {
		fields = append(fields, FieldToProto(subField))
	}
	return fields
}

func FieldToProto(fieldSchema *bigquery.FieldSchema) (field *sharedv1beta1.Field) {
	fieldType, ok := Types.Find(string(fieldSchema.Type))
	if !ok {
		fieldType = v1beta1.NewDataType(fieldSchema.Type)
	}

	return v1beta1.NewField(
		fieldSchema.Name,
		fieldType,
		v1beta1.WithFieldDescription(fieldSchema.Description),
		v1beta1.WithFieldNullable(!fieldSchema.Repeated && !fieldSchema.Required),
		v1beta1.WithFieldRepeated(fieldSchema.Repeated),
		v1beta1.WithFields(FieldsToProto(fieldSchema.Schema)...),
	)
}

func FieldsFromProto(fields v1beta1.Fields) bigquery.Schema {
	var schema bigquery.Schema
	for column := range slices.Values(fields) {
		schema = append(schema, FieldFromProto(column))
	}
	return schema
}

func FieldFromProto(field *sharedv1beta1.Field) *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name:        field.Name,
		Type:        bigquery.FieldType(field.Type.NativeType),
		Description: field.Description,
		Required:    !field.Repeated && !field.Nullable,
		Repeated:    field.Repeated,
		Schema:      FieldsFromProto(field.Fields),
	}
}
