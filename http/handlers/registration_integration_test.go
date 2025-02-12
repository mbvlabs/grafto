//go:build integration
// +build integration

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/routes"
	"github.com/mbvlabs/grafto/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (psql.Postgres, func()) {
	ctx := context.Background()
	postgres, embedded, err := psql.NewPostgresTest(ctx)
	require.NoError(t, err)

	err = postgres.Pool.Ping(ctx)
	require.NoError(t, err)

	cleanup := func() {
		err := embedded.Stop()
		require.NoError(t, err)
	}

	return postgres, cleanup
}

func TestStoreUser(t *testing.T) {
	ctx := context.Background()
	postgres, cleanup := setupTest(t)
	defer cleanup()

	emailSvc := services.NewEmail()

	cacheBuilder, err := otter.NewBuilder[string, string](20)
	require.NoError(t, err)

	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
	require.NoError(t, err)

	handlers := handlers.NewHandlers(postgres, pageCacher, emailSvc)

	routes := routes.NewRoutes(handlers, nil)
	router, ctx := routes.SetupRoutes(ctx)

	tests := []struct {
		name           string
		payload        url.Values
		expectedStatus int
		expectedError  error
	}{
		{
			name: "successful registration",
			payload: url.Values{
				"email":            {"test@example.com"},
				"password":         {"password123"},
				"confirm_password": {"password123"},
			},
			expectedStatus: http.StatusOK,
			expectedError:  pgx.ErrNoRows,
		},
		{
			name: "mismatched passwords",
			payload: url.Values{
				"email":            {"test@example.com"},
				"password":         {"password123"},
				"confirm_password": {"different"},
			},
			expectedStatus: http.StatusOK,
			expectedError:  pgx.ErrNoRows,
		},
		{
			name: "invalid email",
			payload: url.Values{
				"email":            {"notanemail"},
				"password":         {"password123"},
				"confirm_password": {"password123"},
			},
			expectedStatus: http.StatusOK,
			expectedError:  pgx.ErrNoRows,
		},
		{
			name: "empty password",
			payload: url.Values{
				"email":            {"test@example.com"},
				"password":         {""},
				"confirm_password": {""},
			},
			expectedStatus: http.StatusOK,
			expectedError:  pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"http://localhost:8080/register",
				strings.NewReader(tt.payload.Encode()),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := router.NewContext(req, rec)

			err := handlers.Registration.StoreUser(c)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			_, err = models.GetUserByEmail(
				ctx,
				tt.payload.Get("email"),
				postgres.Pool,
			)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}
