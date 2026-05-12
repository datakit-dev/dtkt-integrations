//go:build ignore

package main

import (
	"log"
	"log/slog"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	opts := []entc.Option{
		entc.FeatureNames(
			"entql",
			"schema/snapshot",
			"sql/lock",
			"sql/upsert",
		),
		entc.Dependency(
			entc.DependencyName("Logger"),
			entc.DependencyType(&slog.Logger{}),
		),
	}

	if err := entc.Generate("./pkg/db/schema", &gen.Config{
		Target:  "./pkg/db/ent",
		Schema:  "github.com/datakit-dev/dtkt-integrations/browser/pkg/db/schema",
		Package: "github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent",
	}, opts...); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
