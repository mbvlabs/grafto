package middleware

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
)

func LoginRateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	rateLimitCacheBuilder, err := otter.NewBuilder[string, int32](10_000)
	if err != nil {
		slog.ErrorContext(context.Background(), "failed to create rate limit cache builder", "error", err)

		return func(c echo.Context) error {
			return next(c)
		}
	}

	rateLimiter, err := rateLimitCacheBuilder.WithTTL(10 * time.Minute).Build()
	if err != nil {
		slog.ErrorContext(context.Background(), "failed to build rate limiter cache", "error", err)
		return func(c echo.Context) error {
			return next(c)
		}
	}

	return func(c echo.Context) (err error) {
		ip := c.RealIP()

		hits, found := rateLimiter.Get(ip)
		if !found {
			if ok := rateLimiter.Set(ip, 1); !ok {
				return next(c)
			}
		}
		if hits <= 5 {
			if ok := rateLimiter.Set(ip, hits+1); !ok {
				return next(c)
			}
		}

		if hits > 5 {
			c.Response().
				Header().
				Set("HX-Retarget", "div[id='login-flag']")
			c.Response().
				Header().
				Set("HX-Reswap", "outerHTML")

			return errors.New("")
		}

		return next(c)
	}
}
