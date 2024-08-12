package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/config"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/psql/database"
	"github.com/riverqueue/river"
)

type Base struct {
	cfg         config.Config
	db          *database.Queries
	flashStore  FlashStorage
	queueClient *river.Client[pgx.Tx]
	tracer      telemetry.Tracer
}

func NewDependencies(
	cfg config.Config,
	db *database.Queries,
	flashStore FlashStorage,
	queueClient *river.Client[pgx.Tx],
	tracer telemetry.Tracer,
) Base {
	return Base{
		cfg,
		db,
		flashStore,
		queueClient,
		tracer,
	}
}

func (bd Base) RedirectHx(w http.ResponseWriter, url string) error {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusSeeOther)

	return nil
}

func (bd Base) Redirect(w http.ResponseWriter, r *http.Request, url string) error {
	http.Redirect(w, r, url, http.StatusSeeOther)

	return nil
}

func (bd Base) InternalError(ctx echo.Context) error {
	return ctx.HTML(
		http.StatusOK,
		"<h2>An unrecoverable error occurred. Please click <a href='/'>here</a></h2>",
	)
}
