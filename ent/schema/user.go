package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.String("email").Unique().NotEmpty().MaxLen(255),
		field.String("password_hash").Sensitive().NotEmpty(),
		field.String("name").NotEmpty().MinLen(2).MaxLen(255),
		field.String("avatar_url").Optional().Nillable().MaxLen(500),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("last_login_at").Optional().Nillable(),
		field.Bool("is_active").Default(true),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("notebooks", Notebook.Type),
		edge.To("shared_notebooks", NotebookShare.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("created_at"),
	}
}
