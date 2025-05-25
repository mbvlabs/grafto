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
		return fmt.Errorf("Loki returned error status: %d", resp.StatusCode)
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
