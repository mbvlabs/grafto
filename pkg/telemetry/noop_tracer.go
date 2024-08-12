package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

type NoopSpan struct {
	embedded.Span
}

// AddEvent implements trace.Span.
func (n NoopSpan) AddEvent(name string, options ...trace.EventOption) {
}

// AddLink implements trace.Span.
func (n NoopSpan) AddLink(link trace.Link) {
}

// End implements trace.Span.
func (n NoopSpan) End(options ...trace.SpanEndOption) {
}

// IsRecording implements trace.Span.
func (n NoopSpan) IsRecording() bool {
	return true
}

// RecordError implements trace.Span.
func (n NoopSpan) RecordError(err error, options ...trace.EventOption) {
}

// SetAttributes implements trace.Span.
func (n NoopSpan) SetAttributes(kv ...attribute.KeyValue) {
}

// SetName implements trace.Span.
func (n NoopSpan) SetName(name string) {
}

// SetStatus implements trace.Span.
func (n NoopSpan) SetStatus(code codes.Code, description string) {
}

// SpanContext implements trace.Span.
func (n NoopSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{})
}

// TracerProvider implements trace.Span.
func (n NoopSpan) TracerProvider() trace.TracerProvider {
	return tracesdk.NewTracerProvider()
}

var _ trace.Span = new(NoopSpan)

type NoopTracer struct {
	embedded.Tracer
}

// Start implements trace.Tracer.
func (n NoopTracer) Start(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return context.Background(), NoopSpan{}
}

var _ trace.Tracer = new(NoopTracer)
