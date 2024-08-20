package main

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mbv-labs/grafto/config"
	"github.com/mbv-labs/grafto/psql"
	"github.com/mbv-labs/grafto/psql/bun"

	b "github.com/uptrace/bun"
)

func main() {
	cfg := config.NewConfig()

	conn, err := psql.CreatePooledConnection(
		context.Background(),
		cfg.GetDatabaseURL(),
	)
	if err != nil {
		panic(err)
	}

	db := bun.Setup(conn)

	// var users []bun.User
	// err = db.NewSelect().Model(&users).Scan(context.Background())
	// if err != nil {
	// 	panic(err)
	// }

	var usersMap map[string]any
	err = db.NewSelect().
		Model(&bun.User{}).
		Column("id", "mail").
		Limit(1).
		Scan(context.Background(), &usersMap)
	if err != nil {
		panic(err)
	}

	slog.Info("this is users", "users", usersMap)

	user := &bun.User{}
	qb := db.NewSelect().Model(user).QueryBuilder()
	id := uuid.MustParse("cae4dc8b-51bf-45c2-9e64-5953ff6955ce")
	qb = whereFilter(qb, id)

	q, ok := qb.Unwrap().(*b.SelectQuery)
	if !ok {
		panic(err)
	}

	err = q.Scan(context.Background())
	if err != nil {
		panic(err)
	}

	slog.Info("this is user with where", "user", user)
}

func whereFilter(qb b.QueryBuilder, id uuid.UUID) b.QueryBuilder {
	return qb.Where("id=?", id)
}
