package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type NotebookShare struct {
	ent.Schema
}

func (NotebookShare) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("notebook_id", uuid.UUID{}),
		field.UUID("shared_with_user_id", uuid.UUID{}),
		field.Enum("permission").Values("view", "edit", "admin").Default("view"),
		field.UUID("shared_by_user_id", uuid.UUID{}),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (NotebookShare) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("notebook", Notebook.Type).Ref("shares").
			Field("notebook_id").Unique().Required(),
		edge.From("shared_with", User.Type).Ref("shared_notebooks").
			Field("shared_with_user_id").Unique().Required(),
	}
}

func (NotebookShare) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("notebook_id", "shared_with_user_id").Unique(),
		index.Fields("notebook_id"),
		index.Fields("shared_with_user_id"),
	}
}
