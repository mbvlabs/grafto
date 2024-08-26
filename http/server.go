package http

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
)

type Server struct {
	router *echo.Echo
	host   string
	port   string
	cfg    config.Config
	srv    *http.Server
}

func NewServer(
	router *echo.Echo,
	cfg config.Config,
) Server {
	host := cfg.ServerHost
	port := cfg.ServerPort

	srv := &http.Server{
		Addr: fmt.Sprintf("%v:%v", host, port),
		Handler: csrf.Protect(
			[]byte(
				cfg.CsrfToken,
			),
			csrf.Secure(cfg.Environment == config.PROD_ENVIRONMENT),
			csrf.Path("/"),
		)(
			router,
		),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return Server{
		router,
		host,
		port,
		cfg,
		srv,
	}
}

func (s *Server) Start() {
	slog.Info("starting server on", "host", s.host, "port", s.port)

	// Start server
	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()

	toCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Print("initiating shutdown")
	err := s.srv.Shutdown(toCtx)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("shutdown complete")
}
