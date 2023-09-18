package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Mail           string
	MailVerifiedAt time.Time
}

type NewUser struct {
	ConfirmPassword string
	Name            string
	Mail            string
	Password        string
}
