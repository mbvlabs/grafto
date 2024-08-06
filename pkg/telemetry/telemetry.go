package telemetry

import (
	"log/slog"
	"os"
	"time"

	"github.com/grafana/loki-client-go/loki"
	"github.com/lmittmann/tint"
	"github.com/mbv-labs/grafto/config"
	slogloki "github.com/samber/slog-loki/v3"
)

func NewTelemetry(cfg config.Config, release, service string) {
	switch cfg.Environment {
	case config.PROD_ENVIRONMENT:
		logger := productionLogger(cfg.SinkURL, cfg.TenantID, cfg.Environment, release, service)
		slog.SetDefault(logger)
	case config.DEV_ENVIRONMENT:
		logger := developmentLogger()
		slog.SetDefault(logger)
	default:
		logger := developmentLogger()
		slog.SetDefault(logger)
	}
}

func productionLogger(url, tenantID, release, env, service string) *slog.Logger {
	config, _ := loki.NewDefaultConfig(url)
	config.TenantID = tenantID
	client, err := loki.New(config)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slogloki.Option{Level: slog.LevelInfo, Client: client}.NewLokiHandler())
	logger = logger.
		With("environment", env).
		With("release", release).
		With("service", service)

	return logger
}

func developmentLogger() *slog.Logger {
	return slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		}),
	)
}
