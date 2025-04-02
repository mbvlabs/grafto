package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"golang.org/x/sync/errgroup"
)

type Http struct {
	router *echo.Echo
	host   string
	port   string
	srv    *http.Server
}

func NewHttp(
	ctx context.Context,
	router *echo.Echo,
) Http {
	port := config.Cfg.ServerPort
	host := config.Cfg.ServerHost

	srv := &http.Server{
		Addr: fmt.Sprintf("%v:%v", host, port),
		Handler: func(handler http.Handler) http.Handler {
			return http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/api") ||
						strings.HasPrefix(r.URL.Path, "/river") {

						handler.ServeHTTP(w, r)
						return
					}
					csrf.Protect(
						[]byte(
							config.Cfg.CsrfToken,
						),
						csrf.Secure(
							config.Cfg.Environment == config.PROD_ENVIRONMENT,
						),
						csrf.Path("/"),
					)(handler).ServeHTTP(w, r)
				},
			)
		}(router),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 5 * time.Second,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
	}

	return Http{
		router,
		host,
		port,
		srv,
	}
}

func (s *Http) Start(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	// Start server
	eg.Go(func() error {
		slog.Info("starting server on", "host", s.host, "port", s.port)
		if err := s.srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Handle shutdown on context cancellation
	eg.Go(func() error {
		<-egCtx.Done()
		slog.Info("initiating graceful shutdown")
		shutdownCtx, cancel := context.WithTimeout(
			ctx,
			10*time.Second,
		)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	})

	// Wait for either server error or successful shutdown
	if err := eg.Wait(); err != nil {
		slog.Info("wait error", "e", err)
		return err
	}

	return nil
}
