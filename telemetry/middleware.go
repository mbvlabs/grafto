package telemetry

import (
	"context"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareConfig holds configuration for telemetry middleware
type MiddlewareConfig struct {
	Skipper        func(c echo.Context) bool
	TracerProvider trace.TracerProvider
	Metrics        *ApplicationMetrics
	Logger         *Logger
}

// DefaultMiddlewareConfig provides default configuration
var DefaultMiddlewareConfig = MiddlewareConfig{}

// Middleware returns Echo middleware that provides comprehensive telemetry
func Middleware(cfg ...MiddlewareConfig) echo.MiddlewareFunc {
	config := DefaultMiddlewareConfig
	if len(cfg) > 0 {
		config = cfg[0]
	}

	config.Skipper = func(c echo.Context) bool {
		path := c.Request().URL.Path

		return strings.Contains(path, "/assets/")
	}

	// Use the OpenTelemetry Echo middleware as the base
	otelMiddleware := otelecho.Middleware("grafto")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			ctx := c.Request().Context()

			// Apply OpenTelemetry middleware
			otelNext := otelMiddleware(func(c echo.Context) error {
				return next(c)
			})

			err := otelNext(c)

			// Record metrics after request completion
			if config.Metrics != nil {
				duration := time.Since(start).Seconds()
				statusCode := c.Response().Status

				// HTTP request metrics
				attrs := []attribute.KeyValue{
					attribute.String("method", c.Request().Method),
					attribute.String("route", c.Path()),
					attribute.Int("status_code", statusCode),
				}

				config.Metrics.HTTPRequestCount.Add(
					ctx,
					1,
					metric.WithAttributes(attrs...),
				)
				config.Metrics.HTTPRequestDuration.Record(
					ctx,
					duration,
					metric.WithAttributes(attrs...),
				)
			}

			// Enhanced logging with trace context
			if config.Logger != nil {
				statusCode := c.Response().Status
				duration := time.Since(start)

				logLevel := config.Logger.Info
				if statusCode >= 400 {
					logLevel = config.Logger.Error
				}

				logLevel("HTTP request completed",
					"method", c.Request().Method,
					"path", c.Request().URL.Path,
					"status", statusCode,
					"duration", duration,
					"remote_addr", c.RealIP(),
					"user_agent", c.Request().UserAgent(),
				)
			}

			return err
		}
	}
}

// RequestIDMiddleware adds a request ID to the context and response headers
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get trace ID from current span as request ID
			spanCtx := trace.SpanContextFromContext(c.Request().Context())
			if spanCtx.IsValid() {
				requestID := spanCtx.TraceID().String()
				c.Response().Header().Set("X-Request-ID", requestID)
				c.Set("request_id", requestID)
			}

			return next(c)
		}
	}
}

// RecordAuthMetrics records authentication-related metrics
func RecordAuthMetrics(
	ctx context.Context,
	metrics *ApplicationMetrics,
	success bool,
	authType string,
) {
	if metrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("auth_type", authType),
		attribute.Bool("success", success),
	}

	metrics.AuthAttempts.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordRegistrationMetrics records user registration metrics
func RecordRegistrationMetrics(
	ctx context.Context,
	metrics *ApplicationMetrics,
	success bool,
	source string,
) {
	if metrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("source", source),
		attribute.Bool("success", success),
	}

	metrics.RegistrationCount.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordEmailMetrics records email sending metrics
func RecordEmailMetrics(
	ctx context.Context,
	metrics *ApplicationMetrics,
	emailType string,
	success bool,
) {
	if metrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("email_type", emailType),
		attribute.Bool("success", success),
	}

	metrics.EmailsSent.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordBackgroundJobMetrics records background job processing metrics
func RecordBackgroundJobMetrics(
	ctx context.Context,
	metrics *ApplicationMetrics,
	jobType string,
	success bool,
	duration time.Duration,
) {
	if metrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("job_type", jobType),
		attribute.Bool("success", success),
		attribute.String("duration", duration.String()),
	}

	metrics.BackgroundJobs.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// UpdateDBConnectionMetrics updates database connection pool metrics
func UpdateDBConnectionMetrics(
	ctx context.Context,
	metrics *ApplicationMetrics,
	active, idle int,
) {
	if metrics == nil {
		return
	}

	total := active + idle
	metrics.DBConnections.Add(ctx, int64(total), metric.WithAttributes(
		attribute.String("pool_type", "total"),
	))
}
