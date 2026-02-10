package database

import (
	"context"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"gladiator/ent"
)

// NewEntClient creates an Ent client for the given PostgreSQL DSN and runs auto-migration.
func NewEntClient(databaseURL string) (*ent.Client, error) {
	drv, err := entsql.Open(dialect.Postgres, databaseURL)
	if err != nil {
		return nil, err
	}
	client := ent.NewClient(ent.Driver(drv))
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
