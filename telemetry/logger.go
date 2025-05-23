package telemetry

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"go.opentelemetry.io/otel/trace"
)

// Logger is an enhanced slog.Logger that automatically includes trace context
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new enhanced logger with trace context support
func NewLogger(isDevelopment bool) *Logger {
	var handler slog.Handler
	var level slog.Level

	if isDevelopment {
		level = slog.LevelDebug
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: "15:04:05",
			AddSource:  true,
		})
	}
	if !isDevelopment {
		level = slog.LevelInfo
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: "2006-01-02T15:04:05.000Z07:00",
			AddSource:  false,
		})
	}

	// Wrap the handler to include trace context
	traceHandler := &traceContextHandler{handler: handler}

	return &Logger{
		Logger: slog.New(traceHandler),
	}
}

// NewLoggerWithWriter creates a new enhanced logger with a custom writer
func NewLoggerWithWriter(w io.Writer, isDevelopment bool) *Logger {
	var handler slog.Handler
	var level slog.Level

	if isDevelopment {
		level = slog.LevelDebug
		handler = tint.NewHandler(w, &tint.Options{
			Level:      level,
			TimeFormat: "15:04:05",
			AddSource:  true,
		})
	} else {
		level = slog.LevelError
		handler = tint.NewHandler(w, &tint.Options{
			Level:      level,
			TimeFormat: "2006-01-02T15:04:05.000Z07:00",
			AddSource:  false,
		})
	}

	// Wrap the handler to include trace context
	traceHandler := &traceContextHandler{handler: handler}

	return &Logger{
		Logger: slog.New(traceHandler),
	}
}

// WithContext returns a logger that includes trace context from the given context
func (l *Logger) WithContext(ctx context.Context) *slog.Logger {
	attrs := traceAttrsFromContext(ctx)
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	return l.With(args...)
}

// traceContextHandler wraps an slog.Handler to automatically include trace context
type traceContextHandler struct {
	handler slog.Handler
}

func (h *traceContextHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *traceContextHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	// Add trace context attributes to the record
	traceAttrs := traceAttrsFromContext(ctx)
	for _, attr := range traceAttrs {
		record.AddAttrs(attr)
	}

	return h.handler.Handle(ctx, record)
}

func (h *traceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceContextHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *traceContextHandler) WithGroup(name string) slog.Handler {
	return &traceContextHandler{
		handler: h.handler.WithGroup(name),
	}
}

// traceAttrsFromContext extracts trace context attributes from the context
func traceAttrsFromContext(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)

		if spanCtx.IsSampled() {
			attrs = append(attrs, slog.Bool("trace_sampled", true))
		}
	}

	return attrs
}
