package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AppContext struct {
	echo.Context
	UserID          uuid.UUID
	UserEmail       string
	IsAuthenticated bool
	IsAdmin         bool
	CurrentPath     string
}
