package main

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models/seeds"
	"github.com/mbvlabs/grafto/psql"
)

func main() {
	pool, err := psql.CreatePooledConnection(context.Background(), config.Cfg.GetDatabaseURL())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		panic(err)
	}

	slog.Info("Starting seed script...")

	builder := seeds.NewUserSeedBuilder()
	builder = builder.WithRandoms(5)
	builder.WithSpecific(map[string]any{
		"ID":      uuid.MustParse("0193e547-cb87-76a7-ac91-e1c838c0c9a5"),
		"IsAdmin": true,
		"Email":   "admin@grafto.com",
	})
	userSeed := builder.Build()
	if err := userSeed.Generate(ctx, tx); err != nil {
		panic(err)
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	slog.Info("Seed script finished")
}
