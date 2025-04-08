//go:build integration
// +build integration

package handlers_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/routes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestDB(
	ctx context.Context,
	t *testing.T,
) (psql.Postgres, func(), func()) {
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

func (m *mockedEmailService) Send(
	ctx context.Context,
	payload clients.EmailPayload,
) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

var emailSvc = new(mockedEmailService)

func setupTestHandlers(
	t *testing.T,
	postgres psql.Postgres,
) handlers.Handlers {
	cacheBuilder, err := otter.NewBuilder[string, templ.Component](20)
	require.NoError(t, err)

	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
	require.NoError(t, err)

	return handlers.NewHandlers(postgres, pageCacher, emailSvc)
}

func setupTestRouter(
	ctx context.Context,
	handlers handlers.Handlers,
) (*echo.Echo, context.Context) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	slog.SetDefault(logger)

	routes := routes.NewRoutes(handlers, nil)
	return routes.SetupRoutes(ctx)
}
