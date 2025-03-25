package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models/seeds"
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

	///nolint:errcheck
	defer tx.Rollback(ctx)

	slog.Info("Starting seed script...")

	seeder := seeds.NewSeeder(pool)
	_, err = seeder.PlantUser(
		ctx,
		seeds.WithUserEmailVerifiedAt(time.Now()),
		seeds.WithUserEmail("aryastark@gmail.com"),
	)
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
	slog.Info("Seed script finished")
}
