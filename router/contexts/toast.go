package contexts

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FlashKey struct{}

func (FlashKey) String() string {
	return "flash_key"
}

type FlashType string

const (
	FlashSuccess FlashType = "success"
	FlashError   FlashType = "error"
	FlashWarning FlashType = "warning"
	FlashInfo    FlashType = "info"
)

type FlashMessage struct {
	echo.Context
	ID        uuid.UUID
	Type      FlashType
	CreatedAt time.Time
	Message   string
}
