package telemetry

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type MiddlewareConfig struct {
	Skipper        func(c echo.Context) bool
	TracerProvider trace.TracerProvider
	MetricProvider metric.MeterProvider
}

// RequestIDMiddleware adds a request ID to the context and response headers
// func RequestIDMiddleware() echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			// Get trace ID from current span as request ID
// 			spanCtx := trace.SpanContextFromContext(c.Request().Context())
// 			if spanCtx.IsValid() {
// 				requestID := spanCtx.TraceID().String()
// 				c.Response().Header().Set("X-Request-ID", requestID)
// 				c.Set("request_id", requestID)
// 			}
//
// 			return next(c)
// 		}
// 	}
// }

// RecordAuthMetrics records authentication-related metrics
// func RecordAuthMetrics(
// 	ctx context.Context,
// 	metrics *ApplicationMetrics,
// 	success bool,
// 	authType string,
// ) {
// 	if metrics == nil {
// 		return
// 	}
//
// 	attrs := []attribute.KeyValue{
// 		attribute.String("auth_type", authType),
// 		attribute.Bool("success", success),
// 	}
//
// 	metrics.AuthAttempts.Add(ctx, 1, metric.WithAttributes(attrs...))
// }

// RecordRegistrationMetrics records user registration metrics
// func RecordRegistrationMetrics(
// 	ctx context.Context,
// 	metrics *ApplicationMetrics,
// 	success bool,
// 	source string,
// ) {
// 	if metrics == nil {
// 		return
// 	}
//
// 	attrs := []attribute.KeyValue{
// 		attribute.String("source", source),
// 		attribute.Bool("success", success),
// 	}
//
// 	metrics.RegistrationCount.Add(ctx, 1, metric.WithAttributes(attrs...))
// }

// RecordEmailMetrics records email sending metrics
// func RecordEmailMetrics(
// 	ctx context.Context,
// 	metrics *ApplicationMetrics,
// 	emailType string,
// 	success bool,
// ) {
// 	if metrics == nil {
// 		return
// 	}
//
// 	attrs := []attribute.KeyValue{
// 		attribute.String("email_type", emailType),
// 		attribute.Bool("success", success),
// 	}
//
// 	metrics.EmailsSent.Add(ctx, 1, metric.WithAttributes(attrs...))
// }
//
// RecordBackgroundJobMetrics records background job processing metrics
// func RecordBackgroundJobMetrics(
// 	ctx context.Context,
// 	metrics *ApplicationMetrics,
// 	jobType string,
// 	success bool,
// 	duration time.Duration,
// ) {
// 	if metrics == nil {
// 		return
// 	}
//
// 	attrs := []attribute.KeyValue{
// 		attribute.String("job_type", jobType),
// 		attribute.Bool("success", success),
// 		attribute.String("duration", duration.String()),
// 	}
//
// 	metrics.BackgroundJobs.Add(ctx, 1, metric.WithAttributes(attrs...))
// }

// UpdateDBConnectionMetrics updates database connection pool metrics
// func UpdateDBConnectionMetrics(
// 	ctx context.Context,
// 	metrics *ApplicationMetrics,
// 	active, idle int,
// ) {
// 	if metrics == nil {
// 		return
// 	}
//
// 	metrics.DBConnections.Add(ctx, int64(active), metric.WithAttributes(
// 		attribute.String("pool_type", "active"),
// 	))
// }

// func spanNameFormatter(c echo.Context) string {
// 	method, path := strings.ToUpper(c.Request().Method), c.Path()
// 	if !slices.Contains([]string{
// 		http.MethodGet, http.MethodHead,
// 		http.MethodPost, http.MethodPut,
// 		http.MethodPatch, http.MethodDelete,
// 		http.MethodConnect, http.MethodOptions,
// 		http.MethodTrace,
// 	}, method) {
// 		method = "HTTP"
// 	}
//
// 	if path != "" {
// 		return method + " " + path
// 	}
//
// 	return method
// }

// OPTION 1

// if cfg.TracerProvider == nil {
// 	cfg.TracerProvider = otel.GetTracerProvider()
// }
//
// tracer := cfg.TracerProvider.Tracer(
// 	"grafto",
// 	oteltrace.WithInstrumentationVersion("0.0.1"),
// )
// // if cfg.Propagators == nil {
// // 	cfg.Propagators = otel.GetTextMapPropagator()
// // }
//
// if cfg.Skipper == nil {
// 	cfg.Skipper = middleware.DefaultSkipper
// }
//
// return func(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		if cfg.Skipper(c) {
// 			return next(c)
// 		}
//
// 		c.Set("grafto-trace-key", tracer)
// 		request := c.Request()
// 		savedCtx := request.Context()
// 		defer func() {
// 			request = request.WithContext(savedCtx)
// 			c.SetRequest(request)
// 		}()
// 		// ctx := cfg.Propagators.Extract(
// 		// 	savedCtx,
// 		// 	propagation.HeaderCarrier(request.Header),
// 		// )
// 		opts := []oteltrace.SpanStartOption{
// 			oteltrace.WithAttributes(
// 			// semconvutil.HTTPServerRequest(
// 			// 	service,
// 			// 	request,
// 			// 	semconvutil.HTTPServerRequestOptions{},
// 			// 	nil,
// 			// )...),
// 			),
// 			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
// 		}
// 		if path := c.Path(); path != "" {
// 			rAttr := semconv.HTTPRoute(path)
// 			opts = append(opts, oteltrace.WithAttributes(rAttr))
// 		}
// 		spanName := spanNameFormatter(c)
//
// 		ctx, span := tracer.Start(c.Request().Context(), spanName, opts...)
// 		defer span.End()
//
// 		// pass the span through the request context
// 		c.SetRequest(request.WithContext(ctx))
//
// 		// serve the request to the next middleware
// 		err := next(c)
// 		if err != nil {
// 			span.SetAttributes(attribute.String("echo.error", err.Error()))
// 			// invokes the registered HTTP error handler
// 			c.Error(err)
// 		}
//
// 		status := c.Response().Status
// 		// span.SetStatus(semconvutil.HTTPServerStatus(status))
// 		if status > 0 {
// 			span.SetAttributes(semconv.HTTPStatusCode(status))
// 		}
//
// 		statusCode := c.Response().Status
// 		// duration := time.Since(start)
//
// 		slog.InfoContext(ctx, "HTTP request completed",
// 			"method", c.Request().Method,
// 			"path", c.Request().URL.Path,
// 			"status", statusCode,
// 			// "duration", duration,
// 			"remote_addr", c.RealIP(),
// 			"user_agent", c.Request().UserAgent(),
// 		)
//
// 		return err
// 	}
// }

// OPTION 2
// if cfg.Skipper == nil {
