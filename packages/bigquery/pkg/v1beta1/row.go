package v1beta1

import (
	"context"
	"slices"

	"cloud.google.com/go/bigquery/storage/managedwriter"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/common"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	writeRowsBatch struct {
		tableID     string
		tableDesc   protoreflect.MessageDescriptor
		writeStream *managedwriter.ManagedStream
		result      *managedwriter.AppendResult
		resp        *catalogv1beta1.WriteTableResponse
		rows        writeRowsData
	}
	writeRowData  []byte
	writeRowsData []writeRowData
)

func NewWriteRow(row *structpb.Struct) (writeRowData, error) {
	return common.MarshalJSON[[]byte](row.AsMap())
}

func (r writeRowData) Bytes() []byte {
	return r
}

func (r writeRowData) Size() int {
	return len(r)
}

func (r writeRowsData) Size() (size int) {
	for row := range slices.Values(r) {
		size += row.Size()
	}
	return size
}

func (r writeRowsData) Write(ctx context.Context, writeStream *managedwriter.ManagedStream, tableDesc protoreflect.MessageDescriptor) (*managedwriter.AppendResult, error) {
	var rows [][]byte

	for row := range slices.Values(r) {
		var msg = dynamicpb.NewMessage(tableDesc)
		if err := protojson.Unmarshal(row.Bytes(), msg); err != nil {
			return nil, err
		}

		b, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}

		rows = append(rows, b)
	}

	return writeStream.AppendRows(ctx, rows)
}
