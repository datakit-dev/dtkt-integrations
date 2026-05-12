package pkg

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"buf.build/go/protovalidate"
	localblobv1beta1 "github.com/datakit-dev/dtkt-integrations/localblob/pkg/proto/localblob/v1beta1"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Integration instance struct
type Instance struct {
	config  *localblobv1beta1.Config
	buckets util.SyncMap[string, *blob.Bucket]
}

// Creates a new instance
func NewInstance(ctx context.Context, config *localblobv1beta1.Config) (*Instance, error) {
	err := protovalidate.Validate(config)
	if err != nil {
		return nil, err
	}

	inst := &Instance{
		config: config,
	}

	for name, root := range config.Roots {
		if !filepath.IsAbs(root) {
			return nil, fmt.Errorf("root %s is not absolute: %s", name, root)
		}

		bucket, err := fileblob.OpenBucket(root, &fileblob.Options{
			CreateDir: config.CreateDir,
			NoTempDir: true,
		})
		if err != nil {
			return nil, err
		}

		inst.buckets.Store(name, bucket)
	}

	return inst, nil
}

func (i *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

// Close is called to release instance resources (e.g. client connections)
func (i *Instance) Close() error {
	var errs []error
	for _, bucket := range i.buckets.Values() {
		errs = append(errs, bucket.Close())
	}
	return errors.Join(errs...)
}
