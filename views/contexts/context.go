package contexts

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AppKey struct{}

func (AppKey) String() string {
	return ""
}

type App struct {
	echo.Context
	UserID          uuid.UUID
	Email           string
	IsAuthenticated bool
	IsAdmin         bool
	CurrentPath     string
	Routes          map[string]string
}

func ExtractApp(ctx context.Context) *App {
	appCtx, ok := ctx.Value(AppKey{}).(*App)
	if !ok {
		return &App{}
	}

	return appCtx
}
