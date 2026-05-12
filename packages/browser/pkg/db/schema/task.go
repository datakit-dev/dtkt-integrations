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
)

var _ browserv1beta.Task

type Task struct {
	ent.Schema
}

func (Task) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.ID{},
		mixin.Timestamps{},
	}
}

func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("url"),
		entadapter.EnumField[browserv1beta.TaskState](
			entadapter.WithName("state"),
		),

		// ExtractionTask:
		// - schema_id is required at task creation.
		field.UUID("schema_id", uuid.UUID{}).Optional().Nillable(),
		// - record_id is required at task creation.
		field.UUID("record_id", uuid.UUID{}).Optional().Nillable(),

		field.Time("complete_time").Optional().Nillable(),
		field.String("extension_id"),
	}
}

func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("schema", ExtractionSchema.Type).Unique().Field("schema_id"),
	}
}

func (Task) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("extension_id"),
		index.Fields("extension_id", "schema_id"),
		index.Fields("extension_id", "record_id"),
	}
}
