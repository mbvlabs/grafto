package entity

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID   uuid.UUID `bun:"id,pk"`
	Name string    `bun:"name,notnull"`
}
