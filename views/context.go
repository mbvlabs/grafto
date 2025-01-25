package views

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AppContextKey struct{}

func (AppContextKey) Value() string {
	return "AppContext"
}

type AppContext struct {
	echo.Context
	UserID          uuid.UUID
	Email           string
	IsAuthenticated bool
	IsAdmin         bool
	CurrentPath     string
}

func extractAppContext(ctx context.Context) *AppContext {
	appCtx, ok := ctx.Value(AppContextKey{}.Value()).(*AppContext)
	if !ok {
		return &AppContext{}
	}
	return appCtx
}
