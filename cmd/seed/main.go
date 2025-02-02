package main

import (
	"context"
	"log/slog"

	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/psql"
)

func main() {
	pool, err := psql.CreatePooledConnection(
		context.Background(),
		config.Cfg.GetDatabaseURL(),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		panic(err)
	}

	defer tx.Rollback(ctx)

	slog.Info("Starting seed script...")

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
	slog.Info("Seed script finished")
}
