package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
)

func StartHttp(
	ctx context.Context,
	router *echo.Echo,
) error {
	port := config.Cfg.ServerPort
	host := config.Cfg.ServerHost

	srv := http.Server{
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

	srvErrors := make(chan error, 1)

	go func() {
		slog.InfoContext(ctx, "api server started", "port", "8080")
		srvErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-srvErrors:
		slog.ErrorContext(ctx, "server error", "error", err)
		return err
	case sig := <-shutdown:
		ctxTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		slog.InfoContext(ctx, "server shutdown initiated", "cause", sig)

		if err := srv.Shutdown(ctxTimeout); err != nil {
			slog.ErrorContext(ctx, "server shutdown failed", "error", err)
		}

		slog.InfoContext(ctx, "server shutdown completed")
	}

	return nil
}
