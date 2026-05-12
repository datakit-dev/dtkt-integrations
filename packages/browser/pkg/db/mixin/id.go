package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type ID struct {
	mixin.Schema
}

func (ID) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.Int64("rowid").Immutable().Optional().Annotations(
			entsql.Skip(),
		),
		// field.UUID("id", uuid.UUID{}).
		// 	Default(func() uuid.UUID {
		// 		uid, err := uuid.NewV7()
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 		return uid
		// 	}).
		// 	Annotations(
		// 		entsql.Default("uuidv7()"),
		// 	),
	}
}
