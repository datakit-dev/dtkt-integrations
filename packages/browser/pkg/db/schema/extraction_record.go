package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/mixin"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/protostoresdk/entadapter"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ browserv1beta.ExtractionRecord

// ExtractionRecord holds a single captured value set conforming to an ExtractionSchema.
// One record is produced per task run.
type ExtractionRecord struct {
	ent.Schema
}

func (ExtractionRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.ID{},
		mixin.Timestamps{},
	}
}

func (ExtractionRecord) Fields() []ent.Field {
	return []ent.Field{
		entadapter.MapField[map[string]*structpb.Value](
			entadapter.WithName("values"),
		),
		entadapter.MapField[map[string]*browserv1beta.ElementCapture](
			entadapter.WithName("captures"),
		),

		field.UUID("schema_id", uuid.UUID{}),
		field.UUID("task_id", uuid.UUID{}),
		field.String("extension_id"),
	}
}

func (ExtractionRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("schema", ExtractionSchema.Type).Required().Ref("records").Field("schema_id").Unique(),
	}
}

func (ExtractionRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("extension_id"),
		index.Fields("extension_id", "schema_id"),
		index.Fields("extension_id", "task_id"),
	}
}
