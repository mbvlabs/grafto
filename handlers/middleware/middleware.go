package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var AuthenticatedSessionName = fmt.Sprintf(
	"ua-%s-%s",
	strings.ToLower(config.Cfg.ProjectName),
	config.Cfg.Environment,
)

const (
	FlashSessionKey     = "flash_messages"
	SessIsAuthenticated = "is_authenticated"
	SessUserID          = "user_id"
	SessUserEmail       = "user_email"
	SessIsAdmin         = "is_admin"
	oneWeekInSeconds    = 604800
)

type MW struct {
	rateLimiter       otter.Cache[string, int32]
	tp                trace.TracerProvider
	httpRequestsTotal metric.Int64Counter
	httpDuration      metric.Float64Histogram
	httpInFlight      metric.Int64UpDownCounter
}

func New(tp trace.TracerProvider) (MW, error) {
	rateLimitCacheBuilder, err := otter.NewBuilder[string, int32](10_000)
	if err != nil {
		return MW{}, err
	}

	rateLimit, err := rateLimitCacheBuilder.WithTTL(10 * time.Minute).Build()
	if err != nil {
		return MW{}, err
	}

	// New comprehensive HTTP metrics
	httpRequestsTotal, err := telemetry.HTTPRequestsTotal()
	if err != nil {
		return MW{}, err
	}

	httpDuration, err := telemetry.HTTPRequestDuration()
	if err != nil {
		return MW{}, err
	}

	httpInFlight, err := telemetry.HTTPRequestsInFlight()
	if err != nil {
		return MW{}, err
	}

	return MW{
		rateLimit,
		tp,
		httpRequestsTotal,
		httpDuration,
		httpInFlight,
	}, nil
}
