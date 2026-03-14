package oapigen_test

import (
	"testing"

	"github.com/datakit-dev/dtkt-integrations/openai/pkg/oapigen"
)

func TestOpenAPISpec(t *testing.T) {
	json := oapigen.OpenAPISpec
	if json == nil {
		t.Error("OpenAPI() returned nil")
	} else if len(json) == 0 {
		t.Error("OpenAPI() returned an empty JSON")
	}
}
