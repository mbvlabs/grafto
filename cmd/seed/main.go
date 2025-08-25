package main

import (
	"context"
	"io/fs"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models/seeds"
	"github.com/mbvlabs/grafto/psql"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func main() {
	ctx := context.Background()
	slog.InfoContext(ctx, "Starting seed script...")
	pool, err := psql.CreatePooledConnection(
		context.Background(),
		config.Cfg.DB.GetDatabaseURL(),
	)
	if err != nil {
		panic(err)
	}

	if err := resetDatabase(ctx, pool); err != nil {
		panic(err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		panic(err)
	}

	///nolint:errcheck
	defer tx.Rollback(ctx)

	slog.InfoContext(ctx, "Starting seed script...")

	seeder := seeds.NewSeeder(pool)
	_, err = seeder.PlantUser(
		ctx,
		seeds.WithUserEmailVerifiedAt(time.Now()),
		seeds.WithUserEmail("aryastark@gmail.com"),
		seeds.WithUserIsAdmin(true),
	)
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "Seed script finished")
}

func resetDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	slog.InfoContext(ctx, "Resetting database...")
	gooseLock, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}

	fsys, err := fs.Sub(psql.Migrations, "migrations")
	if err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pool)

	gooseProvider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		fsys,
		goose.WithVerbose(false),
		goose.WithSessionLocker(gooseLock),
	)
	if err != nil {
		return err
	}

	_, err = gooseProvider.DownTo(ctx, 0)
	if err != nil {
		return err
	}
	_, err = gooseProvider.Up(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Database reset finished")
	return nil
}
