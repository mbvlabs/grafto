//go:build integration
// +build integration

package controllers_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/controllers"
	"github.com/mbvlabs/grafto/pkg/clients"
	"github.com/mbvlabs/grafto/pkg/telemetry"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router"
	"github.com/mbvlabs/grafto/services"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestDB(
	ctx context.Context,
	t *testing.T,
) (psql.Postgres, func(context.Context), func()) {
	testPsql, err := psql.NewPostgresTest(ctx)
	require.NoError(t, err)

	err = testPsql.Psql.Pool.Ping(ctx)
	require.NoError(t, err)

	stopEmbedded := func() {
		err := testPsql.EmbeddedPsql.Stop()
		require.NoError(t, err)
	}

	return testPsql.Psql, testPsql.CleanupFunc, stopEmbedded
}

type mockedEmailService struct {
	mock.Mock
}

func (m *mockedEmailService) SendTransaction(
	ctx context.Context,
	payload clients.EmailPayload,
) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func setupTestControllers(
	t *testing.T,
	postgres psql.Postgres,
	emailSvc services.EmailSender,
) controllers.Controllers {
	cacheBuilder, err := otter.NewBuilder[string, templ.Component](20)
	require.NoError(t, err)

	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
	require.NoError(t, err)

	return controllers.New(postgres, pageCacher, emailSvc)
}

func setupTestRouter(
	t *testing.T,
	controllers controllers.Controllers,
	// mw middleware.MW,
) *echo.Echo {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	tp, err := telemetry.NewTraceProvider(
		context.Background(),
		nil,
		&telemetry.NoopTraceExporter{},
		0.0,
	)

	require.NoError(t, err, "new trace exporter returned error ")

	router, err := router.New(controllers, nil, tp)
	require.NoError(t, err, "new router returned error ")

	return router.SetupRoutes()
}

type (
	config struct {
		Skipper echomw.Skipper
		Store   sessions.Store
	}
)

const (
	key = "_session_store"
)

var testDefaultConfig = config{
	Skipper: echomw.DefaultSkipper,
}

func testCookieStore(store sessions.Store) echo.MiddlewareFunc {
	c := testDefaultConfig
	c.Store = store
	return testMiddlewareCookieStoreWithConfig(c)
}

func testMiddlewareCookieStoreWithConfig(config config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = testDefaultConfig.Skipper
	}
	if config.Store == nil {
		panic("echo: session middleware requires store")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			c.Set(key, config.Store)
			return next(c)
		}
	}
}
