package lib

import (
	"context"

	"cloud.google.com/go/bigquery"
	bigqueryv1beta "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/proto/integration/bigquery/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

type Client struct {
	*bigquery.Client
	Config    *bigqueryv1beta.Config
	CredsJson []byte
	Options   []option.ClientOption
}

func NewClient(ctx context.Context, config *bigqueryv1beta.Config, opts ...option.ClientOption) (*Client, error) {
	var credsJson []byte
	if config.CredentialsJson != nil {
		b, err := protojson.Marshal(config.CredentialsJson)
		if err != nil {
			return nil, err
		}

		credsJson = b
		opts = append(opts, option.WithCredentialsJSON(b))
	}

	opts = append(opts, option.WithLogger(log.FromCtx(ctx)))

	// if projectId == "" {
	// 	projectId = bigquery.DetectProjectID
	// }

	client, err := bigquery.NewClient(ctx, config.ProjectId, opts...)
	if err != nil {
		return nil, err
	}

	client.Location = config.Location

	query := client.Query("SELECT 1")
	query.DryRun = true
	_, err = query.Run(ctx)
	if err != nil {
		return nil, err
	}

	err = client.EnableStorageReadClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:    client,
		Config:    config,
		CredsJson: credsJson,
		Options:   opts,
	}, nil
}
