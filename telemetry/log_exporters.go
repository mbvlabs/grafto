package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lmittmann/tint"
)

type StdoutExporter struct {
	LogLevel   slog.Level
	WithTraces bool
}

// GetSlogHandler implements LogExporter.
func (s *StdoutExporter) GetSlogHandler(
	ctx context.Context,
) (slog.Handler, error) {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      s.LogLevel,
		TimeFormat: "15:04:05",
		AddSource:  true,
	})

	if s.WithTraces {
		return &traceLogHandler{handler: handler}, nil
	}

	return handler, nil
}

// Name implements LogExporter.
func (s *StdoutExporter) Name() string {
	return "stdout"
}

// Shutdown implements LogExporter.
func (s *StdoutExporter) Shutdown(ctx context.Context) error {
	return nil
}

var _ LogExporter = new(StdoutExporter)

type LokiExporter struct {
	LogLevel   slog.Level
	WithTraces bool
	URL        string
	Labels     map[string]string
	httpClient *http.Client
}

// LokiPushRequest represents the Loki push API request format
type LokiPushRequest struct {
	Streams []LokiStream `json:"streams"`
}

type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// GetSlogHandler implements LogExporter.
func (l *LokiExporter) GetSlogHandler(
	ctx context.Context,
) (slog.Handler, error) {
	if l.httpClient == nil {
		l.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	handler := &lokiHandler{
		url:        l.URL,
		httpClient: l.httpClient,
		logLevel:   l.LogLevel,
		labels:     l.Labels,
	}

	if l.WithTraces {
		return &traceLogHandler{handler: handler}, nil
	}

	return handler, nil
}

// Name implements LogExporter.
func (l *LokiExporter) Name() string {
	return "loki"
}

// Shutdown implements LogExporter.
func (l *LokiExporter) Shutdown(ctx context.Context) error {
	return nil
}

var _ LogExporter = new(LokiExporter)

// lokiHandler implements slog.Handler for Loki HTTP API
type lokiHandler struct {
	url        string
	httpClient *http.Client
	logLevel   slog.Level
	labels     map[string]string
	attrs      []slog.Attr
	groups     []string
}

func (h *lokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.logLevel
}

func (h *lokiHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	// Build labels
	labels := make(map[string]string)
	for k, v := range h.labels {
		labels[k] = v
	}
	labels["level"] = record.Level.String()

	// Build log line with attributes
	logLine := record.Message

	// Add existing attributes
	for _, attr := range h.attrs {
		logLine += fmt.Sprintf(" %s=%v", attr.Key, attr.Value)
	}

	// Add record attributes
	record.Attrs(func(attr slog.Attr) bool {
		logLine += fmt.Sprintf(" %s=%v", attr.Key, attr.Value)
		return true
	})

	// Create Loki push request
	timestamp := strconv.FormatInt(record.Time.UnixNano(), 10)
	pushReq := LokiPushRequest{
		Streams: []LokiStream{
			{
				Stream: labels,
				Values: [][]string{
					{timestamp, logLine},
				},
			},
		},
	}

	// Send to Loki via HTTP
	return h.sendToLoki(ctx, pushReq)
}

func (h *lokiHandler) sendToLoki(
	ctx context.Context,
	req LokiPushRequest,
) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		h.url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send log to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("loki returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (h *lokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &lokiHandler{
		url:        h.url,
		httpClient: h.httpClient,
		logLevel:   h.logLevel,
		labels:     h.labels,
		attrs:      newAttrs,
		groups:     h.groups,
	}
}

func (h *lokiHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &lokiHandler{
		url:        h.url,
		httpClient: h.httpClient,
		logLevel:   h.logLevel,
		labels:     h.labels,
		attrs:      h.attrs,
		groups:     newGroups,
	}
}

// BetterStackLogExporter implements LogExporter for BetterStack
type BetterStackLogExporter struct {
	Endpoint    string
	SourceToken string
	LogLevel    slog.Level
	WithTraces  bool
	exporter    *betterStackLogHandler
}

func NewBetterStackLogExporter(
	endpoint, sourceToken string,
) *BetterStackLogExporter {
	return &BetterStackLogExporter{
		Endpoint:    endpoint,
		SourceToken: sourceToken,
		LogLevel:    slog.LevelInfo,
	}
}

func (b *BetterStackLogExporter) Name() string {
	return "betterstack"
}

func (b *BetterStackLogExporter) GetSlogHandler(
	ctx context.Context,
) (slog.Handler, error) {
	if b.exporter == nil {
		b.exporter = &betterStackLogHandler{
			endpoint:   b.Endpoint,
			token:      b.SourceToken,
			logLevel:   b.LogLevel,
			httpClient: &http.Client{Timeout: 10 * time.Second},
		}
	}

	if b.WithTraces {
		return &traceLogHandler{handler: b.exporter}, nil
	}

	return b.exporter, nil
}

func (b *BetterStackLogExporter) Shutdown(ctx context.Context) error {
	return nil
}

var _ LogExporter = new(BetterStackLogExporter)

// betterStackLogHandler implements slog.Handler for BetterStack OTLP HTTP API
type betterStackLogHandler struct {
	endpoint   string
	token      string
	httpClient *http.Client
	logLevel   slog.Level
	attrs      []slog.Attr
	groups     []string
}

// BetterStackOTLPLog represents the OTLP log format for BetterStack
type BetterStackOTLPLog struct {
	ResourceLogs []BetterStackResourceLog `json:"resourceLogs"`
}

type BetterStackResourceLog struct {
	ScopeLogs []BetterStackScopeLog `json:"scopeLogs"`
}

type BetterStackScopeLog struct {
	LogRecords []BetterStackLogRecord `json:"logRecords"`
}

type BetterStackLogRecord struct {
	TimeUnixNano   string                    `json:"timeUnixNano"`
	SeverityNumber int                       `json:"severityNumber"`
	SeverityText   string                    `json:"severityText"`
	Body           BetterStackLogBody        `json:"body"`
	Attributes     []BetterStackLogAttribute `json:"attributes,omitempty"`
}

type BetterStackLogBody struct {
	StringValue string `json:"stringValue"`
}

type BetterStackLogAttribute struct {
	Key   string                    `json:"key"`
	Value BetterStackAttributeValue `json:"value"`
}

type BetterStackAttributeValue struct {
	StringValue string `json:"stringValue"`
}

func (h *betterStackLogHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return level >= h.logLevel
}

func (h *betterStackLogHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	// Build log message with attributes
	logMessage := record.Message
	var attributes []BetterStackLogAttribute

	// Add existing attributes
	for _, attr := range h.attrs {
		attributes = append(attributes, BetterStackLogAttribute{
			Key: attr.Key,
			Value: BetterStackAttributeValue{
				StringValue: attr.Value.String(),
			},
		})
	}

	// Add record attributes
	record.Attrs(func(attr slog.Attr) bool {
		attributes = append(attributes, BetterStackLogAttribute{
			Key: attr.Key,
			Value: BetterStackAttributeValue{
				StringValue: attr.Value.String(),
			},
		})
		return true
	})

	// Map slog level to OTLP severity
	severityNumber := h.mapSlogLevelToOTLP(record.Level)

	// Create OTLP log record
	logRecord := BetterStackLogRecord{
		TimeUnixNano:   fmt.Sprintf("%d", record.Time.UnixNano()),
		SeverityNumber: severityNumber,
		SeverityText:   record.Level.String(),
		Body: BetterStackLogBody{
			StringValue: logMessage,
		},
		Attributes: attributes,
	}

	otlpLog := BetterStackOTLPLog{
		ResourceLogs: []BetterStackResourceLog{
			{
				ScopeLogs: []BetterStackScopeLog{
					{
						LogRecords: []BetterStackLogRecord{logRecord},
					},
				},
			},
		},
	}

	// Send to BetterStack
	return h.sendToBetterStack(ctx, otlpLog)
}

func (h *betterStackLogHandler) mapSlogLevelToOTLP(level slog.Level) int {
	switch level {
	case slog.LevelDebug:
		return 5 // TRACE
	case slog.LevelInfo:
		return 9 // INFO
	case slog.LevelWarn:
		return 13 // WARN
	case slog.LevelError:
		return 17 // ERROR
	default:
		return 9 // INFO
	}
}

func (h *betterStackLogHandler) sendToBetterStack(
	ctx context.Context,
	log BetterStackOTLPLog,
) error {
	jsonData, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		h.endpoint,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send log to BetterStack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf(
			"BetterStack returned error status: %d",
			resp.StatusCode,
		)
	}

	return nil
}

func (h *betterStackLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &betterStackLogHandler{
		endpoint:   h.endpoint,
		token:      h.token,
		httpClient: h.httpClient,
		logLevel:   h.logLevel,
		attrs:      newAttrs,
		groups:     h.groups,
	}
}

func (h *betterStackLogHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &betterStackLogHandler{
		endpoint:   h.endpoint,
		token:      h.token,
		httpClient: h.httpClient,
		logLevel:   h.logLevel,
		attrs:      h.attrs,
		groups:     newGroups,
	}
}
