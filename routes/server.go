package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"log/slog"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/config"
	"github.com/MBvisti/grafto/routes/api"
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
}

func NewServer(
	router *echo.Echo, controllers controllers.Controller, logger *slog.Logger, cfg config.Cfg) Server {
	api := api.NewAPI(router, controllers, logger)
	api.SetupAPIRoutes()

	web := web.NewWeb(router, controllers)
	web.SetupWebRoutes()

	if cfg.App.Environment == "development" {
		router.Debug = true
	}

	router.Static("/static", "static")

	return Server{
		router,
		cfg.App.ServerHost,
		cfg.App.ServerPort,
		api,
		web,
		cfg,
	}
}

func (s *Server) Start() {
	isProduction := s.cfg.App.Environment == "production"

	slog.Info("starting server on", "host", s.host, "port", s.port)
	srv := http.Server{
		Addr: fmt.Sprintf("%v:%v", s.host, s.port),
		Handler: csrf.Protect(
			[]byte(s.cfg.Auth.CsrfToken), csrf.Secure(isProduction))(s.router),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
