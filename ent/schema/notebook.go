package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Notebook struct {
	ent.Schema
}

func (Notebook) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("owner_id", uuid.UUID{}),
		field.String("title").NotEmpty().MinLen(1).MaxLen(255),
		field.Text("description").Optional().Nillable(),
		field.JSON("content", map[string]interface{}{}).
			Default(map[string]interface{}{"cells": []interface{}{}}),
		field.Bool("is_public").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("last_executed_at").Optional().Nillable(),
		field.Int("execution_count").Default(0).NonNegative(),
	}
}

func (Notebook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("notebooks").
			Field("owner_id").Unique().Required(),
		edge.To("shares", NotebookShare.Type).
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

func (Notebook) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "created_at"),
		index.Fields("owner_id", "updated_at"),
		index.Fields("is_public", "created_at"),
		index.Fields("updated_at"),
	}
}
