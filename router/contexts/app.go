package contexts

import (
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
	// TraceID         string
	// SpanID          string
}

// GetTraceContext returns trace information from the request context
// func (a App) GetTraceContext() (traceID, spanID string) {
// 	spanCtx := trace.SpanContextFromContext(a.Request().Context())
// 	if spanCtx.IsValid() {
// 		return spanCtx.TraceID().String(), spanCtx.SpanID().String()
// 	}
// 	return a.TraceID, a.SpanID
// }
