package oapigen

import (
	_ "embed"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/common"
)

//go:embed openapi.json
var openAPIRaw []byte
var OpenAPISpec = common.MustUnmarshalJSON[common.JSONMap](openAPIRaw)
