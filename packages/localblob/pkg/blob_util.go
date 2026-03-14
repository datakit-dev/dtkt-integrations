package pkg

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"
	"gocloud.dev/blob"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type BucketRequest interface {
	GetBucket() string
}

func GetBucket(buckets *util.SyncMap[string, *blob.Bucket], req BucketRequest) (*blob.Bucket, error) {
	if req.GetBucket() == "" {
		return nil, fmt.Errorf("invalid request: bucket is required")
	}

	bucket, ok := buckets.Load(req.GetBucket())
	if !ok {
		return nil, fmt.Errorf("invalid request: bucket not found: %s", req.GetBucket())
	}

	return bucket, nil
}

func GetBlob(ctx context.Context, bucket *blob.Bucket, key string) (*blob.Attributes, error) {
	attrs, err := bucket.Attributes(ctx, key)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// md5Hash := string(attrs.MD5)
	// If MD5 is not set, compute it
	if len(attrs.MD5) == 0 {
		reader, err := bucket.NewReader(ctx, key, nil)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to read blob for MD5 computation: %v", err))
		}
		//nolint:errcheck
		defer reader.Close()

		hash := md5.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to compute MD5: %v", err))
		}
		// md5Hash = hex.EncodeToString(hash.Sum(nil))
		attrs.MD5 = hash.Sum(nil)
	}

	return attrs, nil
}
