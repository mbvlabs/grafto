package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/routes/api"
	"github.com/MBvisti/grafto/routes/middleware"
	"github.com/MBvisti/grafto/services"
	"github.com/gorilla/csrf"

	"github.com/MBvisti/grafto/routes/web"
	"github.com/labstack/echo/v4"
)

type Server struct {
	router *echo.Echo
	host   string
	port   string
	api    api.API
	web    web.Web
	cfg    config.Cfg
	srv    *http.Server
}

func NewServer(
	router *echo.Echo, controllers controllers.Controller, logger *slog.Logger, cfg config.Cfg, services services.Services) Server {
	api := api.NewAPI(router, controllers, logger)
	api.SetupAPIRoutes()

	middleware := middleware.NewMiddleware(services)

	web := web.NewWeb(router, controllers, middleware)
	web.SetupWebRoutes()

	if cfg.App.Environment == "development" {
		router.Debug = true
	}

	router.Static("/static", "static")

	host := cfg.App.ServerHost
	port := cfg.App.ServerPort
	isProduction := cfg.App.Environment == "production"

	srv := &http.Server{
		Addr: fmt.Sprintf("%v:%v", host, port),
		Handler: csrf.Protect(
			[]byte(cfg.Auth.CsrfToken), csrf.Secure(isProduction))(router),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return Server{
		router,
		host,
		port,
		api,
		web,
		cfg,
		srv,
	}
}

func (s *Server) Start() {
	slog.Info("starting server on", "host", s.host, "port", s.port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	// Start server
	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	// ctxWTO, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	//
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
