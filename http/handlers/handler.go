package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/pkg/config"
	"github.com/mbv-labs/grafto/repository/psql/database"
	"github.com/riverqueue/river"
)

type Base struct {
	cfg         config.Cfg
	db          *database.Queries
	flashStore  FlashStorage
	queueClient *river.Client[pgx.Tx]
}

func NewDependencies(
	cfg config.Cfg,
	db *database.Queries,
	flashStore FlashStorage,
	queueClient *river.Client[pgx.Tx],
) Base {
	return Base{
		cfg,
		db,
		flashStore,
		queueClient,
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
	return ctx.JSON(http.StatusOK, "dsadads")
	// from := "/"
	//
	// return views.InternalServerErr(ctx, views.InternalServerErrData{
	// 	FromLocation: from,
	// })
}
