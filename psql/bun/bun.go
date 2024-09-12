package bun

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type User struct {
	bun.BaseModel  `bun:"table:users"`
	ID             uuid.UUID `bun:",pk"`
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	Name           string
	Mail           string
	MailVerifiedAt pgtype.Timestamptz
	Password       string
}

func Setup(pool *pgxpool.Pool) *bun.DB {
	sqldb := stdlib.OpenDBFromPool(pool)
	db := bun.NewDB(sqldb, pgdialect.New())

	return db
}
