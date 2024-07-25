package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Email           string
	EmailVerifiedAt time.Time
}

func (u User) IsVerified() bool {
	return !u.EmailVerifiedAt.IsZero()
}
