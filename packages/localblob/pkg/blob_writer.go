package pkg

import (
	"context"

	blobv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/blob/v1beta1"
	"gocloud.dev/blob"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BlobLineWriter struct {
	*blob.Writer
	bucket       *blob.Bucket
	key          string
	linesWritten int64
}

func (w *BlobLineWriter) Result(ctx context.Context) (*blobv1beta1.WriteBlobLinesResponse_KeyResult, error) {
	attrs, err := GetBlob(ctx, w.bucket, w.key)
	if err != nil {
		return nil, err
	}

	return &blobv1beta1.WriteBlobLinesResponse_KeyResult{
		Key:          w.key,
		Size:         attrs.Size,
		ModTime:      timestamppb.New(attrs.ModTime),
		Etag:         attrs.ETag,
		LinesWritten: w.linesWritten,
	}, nil
}
