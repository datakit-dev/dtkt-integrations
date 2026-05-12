package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type Timestamps struct {
	mixin.Schema
}

func (Timestamps) Fields() []ent.Field {
	return []ent.Field{
		field.Time("create_time").
			Default(time.Now).
			Immutable(),
		field.Time("update_time").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Timestamps) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("create_time"),
		index.Fields("update_time"),
	}
}
