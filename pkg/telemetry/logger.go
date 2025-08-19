package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type LogExporter interface {
	Name() string
	GetSlogHandler(
		ctx context.Context,
	) (slog.Handler, error)

	Shutdown(ctx context.Context) error
}

func NewLogger(
	ctx context.Context,
	exporter LogExporter,
) (*slog.Logger, func(ctx context.Context) error) {
	handler, err := exporter.GetSlogHandler(ctx)
	if err != nil {
		panic(err)
	}

	return slog.New(handler), exporter.Shutdown
}

// traceLogHandler wraps an slog.Handler to automatically include trace context
type traceLogHandler struct {
	handler slog.Handler
}

func (h *traceLogHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *traceLogHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	traceAttrs := traceAttrsFromContext(ctx)
	for _, attr := range traceAttrs {
		record.AddAttrs(attr)
	}

	return h.handler.Handle(ctx, record)
}

func (h *traceLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceLogHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *traceLogHandler) WithGroup(name string) slog.Handler {
	return &traceLogHandler{
		handler: h.handler.WithGroup(name),
	}
}

func traceAttrsFromContext(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	return attrs
}
