package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/mixin"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/protostoresdk/entadapter"
)

var _ browserv1beta.ExtractionSchema

// ExtractionSchema holds a named, reusable set of FieldDefs.
type ExtractionSchema struct {
	ent.Schema
}

func (ExtractionSchema) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.ID{},
		mixin.Timestamps{},
	}
}

func (ExtractionSchema) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Optional().Nillable(),
		entadapter.ListField[[]*browserv1beta.FieldDef](
			entadapter.WithName("fields"),
		),
		field.String("extension_id"),
	}
}

func (ExtractionSchema) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("records", ExtractionRecord.Type),
		edge.From("tasks", Task.Type).Ref("schema"),
	}
}

func (ExtractionSchema) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("extension_id"),
		index.Fields("extension_id", "name"),
	}
}
