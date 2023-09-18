package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

func SetupDatabaseConnection(databaseURL string) *pgx.Conn {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return conn
}
