//go:build integration
// +build integration

package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/models/seeds"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/routes"
	"github.com/mbvlabs/grafto/services"
	"github.com/stretchr/testify/assert"
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

func setupTestHandlers(
	postgres psql.Postgres,
	t *testing.T,
) handlers.Handlers {
	emailSvc := services.NewEmail()

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
	routes := routes.NewRoutes(handlers, nil)
	return routes.SetupRoutes(ctx)
}

func TestStoreUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	handlers := setupTestHandlers(postgres, t)
	router, ctx := setupTestRouter(ctx, handlers)

	tests := []struct {
		name           string
		payload        url.Values
		expectedStatus int
		expectedError  error
	}{
		{
			name: "should register a new user",
			payload: url.Values{
				"email":            {"test@example.com"},
				"password":         {"password123"},
				"confirm_password": {"password123"},
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "should not register new user because mismatched passwords",
			payload: url.Values{
				"email":            {"test1@example.com"},
				"password":         {"password123"},
				"confirm_password": {"different"},
			},
			expectedStatus: http.StatusOK,
			expectedError:  pgx.ErrNoRows,
		},
		{
			name: "should not register new user invalid email",
			payload: url.Values{
				"email":            {"notanemail"},
				"password":         {"password123"},
				"confirm_password": {"password123"},
			},
			expectedStatus: http.StatusOK,
			expectedError:  pgx.ErrNoRows,
		},
		{
			name: "should not register new user empty password",
			payload: url.Values{
				"email":            {"test2@example.com"},
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
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

func TestVerifyEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	handlers := setupTestHandlers(postgres, t)
	router, ctx := setupTestRouter(ctx, handlers)

	seeder := seeds.NewSeeder(postgres.Pool)
	tests := []struct {
		name             string
		email            string
		token            func(ctx context.Context) models.Token
		expectedStatus   int
		expectedVerified bool
	}{
		{
			name: "should validate email",
			token: func(ctx context.Context) models.Token {
				user, err := seeder.PlantUser(ctx)
				if err != nil {
					t.FailNow()
				}

				tkn, err := seeder.PlantToken(
					ctx,
					seeds.WithTokenMeta(models.MetaInformation{
						Resource:   models.ResourceUser,
						ResourceID: user.ID,
						Scope:      models.ScopeEmailVerification,
					}),
				)
				if err != nil {
					t.FailNow()
				}

				return tkn
			},
			expectedStatus:   http.StatusOK,
			expectedVerified: true,
		},
		{
			name: "should not validate email",
			token: func(ctx context.Context) models.Token {
				user, err := seeder.PlantUser(ctx)
				if err != nil {
					t.FailNow()
				}

				tkn, err := seeder.PlantToken(
					ctx,
					seeds.WithTokenExpiration(time.Now().Add(-1*time.Hour)),
					seeds.WithTokenMeta(models.MetaInformation{
						Resource:   models.ResourceUser,
						ResourceID: user.ID,
						Scope:      models.ScopeEmailVerification,
					}),
				)
				if err != nil {
					t.FailNow()
				}

				return tkn
			},
			expectedStatus:   http.StatusOK,
			expectedVerified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.token(ctx)

			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				fmt.Sprintf(
					"http://localhost:8080/verify-email?token=%s",
					token.Hash,
				),
				nil,
			)
			rec := httptest.NewRecorder()
			c := router.NewContext(req, rec)

			err := handlers.Registration.VerifyUserEmail(c)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			usr, err := models.GetUser(
				ctx,
				token.Meta.ResourceID,
				postgres.Pool,
			)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedVerified, !usr.EmailVerifiedAt.IsZero())
		})
	}
}
