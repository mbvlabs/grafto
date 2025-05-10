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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/models/seeds"
	"github.com/mbvlabs/grafto/router/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	testHandlers := setupTestHandlers(t, postgres)
	router, ctx := setupTestRouter(ctx, testHandlers)

	tests := []struct {
		name          string
		payload       url.Values
		expectedError error
	}{
		{
			name: "should register a new user",
			payload: url.Values{
				"email": {
					fmt.Sprintf("%s@gmail.com", uuid.New().String()),
				},
				"password":         {"password123"},
				"confirm_password": {"password123"},
			},
			expectedError: nil,
		},
		{
			name: "should not register new user because mismatched passwords",
			payload: url.Values{
				"email": {
					fmt.Sprintf("%s@gmail.com", uuid.New().String()),
				},
				"password":         {"password123"},
				"confirm_password": {"different"},
			},
			expectedError: pgx.ErrNoRows,
		},
		{
			name: "should not register new user invalid email",
			payload: url.Values{
				"email":            {"notanemail"},
				"password":         {"password123"},
				"confirm_password": {"password123"},
			},
			expectedError: pgx.ErrNoRows,
		},
		{
			name: "should not register new user empty password",
			payload: url.Values{
				"email": {
					fmt.Sprintf("%s@gmail.com", uuid.New().String()),
				},
				"password":         {""},
				"confirm_password": {""},
			},
			expectedError: pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				fmt.Sprintf(
					"http://localhost:8080%s",
					routes.StoreUser.Path,
				),
				strings.NewReader(tt.payload.Encode()),
			)

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			emailSvc.On(
				"SendTransaction",
				mock.Anything,
				mock.MatchedBy(func(payload clients.EmailPayload) bool {
					return true
				}),
				mock.Anything,
			).Return(nil)

			router.ServeHTTP(rec, req)

			_, err := models.GetUserByEmail(
				ctx,
				postgres.Pool,
				tt.payload.Get("email"),
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

	testHandlers := setupTestHandlers(t, postgres)
	router, ctx := setupTestRouter(ctx, testHandlers)

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
					"http://localhost:8080%s?token=%s",
					routes.VerifyEmail.Path,
					token.Value,
				),
				nil,
			)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			usr, err := models.GetUser(
				ctx,
				postgres.Pool,
				token.Meta.ResourceID,
			)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedVerified, !usr.EmailVerifiedAt.IsZero())
		})
	}
}
