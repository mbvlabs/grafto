package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/MBvisti/grafto/controllers"
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
}

func NewServer(
	router *echo.Echo, controllers controllers.Controller, logger *slog.Logger) Server {
	api := api.NewAPI(router, controllers, logger)
	api.SetupAPIRoutes()

	web := web.NewWeb(router, controllers)
	web.SetupWebRoutes()

	host := os.Getenv("SERVER_HOST")
	if host == "" {
		panic("server host env variable empty")
	}

	port := os.Getenv("SERVER_PORT")
	if host == "" {
		panic("server port env variable empty")
	}

	if os.Getenv("ENV") == "development" {
		router.Debug = true
	}

	router.Static("/static", "static")

	return Server{
		router,
		host,
		port,
		api,
		web,
	}
}

func (s *Server) Start() {
	csrf := csrf.Protect(
		[]byte(os.Getenv("CSRF_TOKEN")), csrf.Secure(false), csrf.CookieName("_csrf"), csrf.FieldName("_csrf"))

	slog.Info("starting server on", "host", s.host, "port", s.port)
	srv := http.Server{
		Addr:         fmt.Sprintf("%v:%v", s.host, s.port),
		Handler:      csrf(s.router),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
