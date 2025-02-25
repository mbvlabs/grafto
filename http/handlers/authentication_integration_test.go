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

	"github.com/gorilla/sessions"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/models/seeds"
	"github.com/stretchr/testify/assert"
)

func TestStoreAuthenticatedSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	testHandlers := setupTestHandlers(postgres, t)
	router, ctx := setupTestRouter(ctx, testHandlers)

	seeder := seeds.NewSeeder(postgres.Pool)
	validUser, err := seeder.PlantUser(
		ctx,
		seeds.WithUserEmailVerifiedAt(time.Now()),
		seeds.WithUserEmail("jonsnow@gmail.com"),
	)
	assert.NoError(t, err)
	invalidUser, err := seeder.PlantUser(
		ctx,
		seeds.WithUserEmail("sansastark@gmail.com"),
	)
	assert.NoError(t, err)

	tests := []struct {
		name              string
		payload           url.Values
		expectedToSucceed bool
	}{
		{
			name: "should authenticate user",
			payload: url.Values{
				"email":       {validUser.Email},
				"password":    {"password"},
				"remember_me": {"on"},
			},
			expectedToSucceed: true,
		},
		{
			name: "should not authenticate the user bc password is wrong",
			payload: url.Values{
				"email":       {validUser.Email},
				"password":    {"password123"},
				"remember_me": {"on"},
			},
			expectedToSucceed: false,
		},
		{
			name: "should not authenticate the user bc email not validated",
			payload: url.Values{
				"email":       {invalidUser.Email},
				"password":    {"password"},
				"remember_me": {"on"},
			},
			expectedToSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				fmt.Sprintf("%s/login", config.Cfg.GetFullDomain()),
				strings.NewReader(tt.payload.Encode()),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			c := router.NewContext(req, rec)
			c.Set(
				"_session_store",
				sessions.NewCookieStore(
					[]byte(config.Cfg.SessionEncryptionKey),
				),
			)

			err := testHandlers.Authentication.StoreAuthenticatedSession(c)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)

			cookies := rec.Result().Cookies()
			var authToken string
			for _, cookie := range cookies {
				if cookie.Name == handlers.AuthenticatedSessionName {
					authToken = cookie.Value
					break
				}
			}

			if tt.expectedToSucceed {
				assert.NotEmpty(
					t,
					authToken,
					"Auth token should not be empty",
				)
			}
			if !tt.expectedToSucceed {
				assert.Empty(
					t,
					authToken,
					"Auth token should be empty",
				)
			}
		})
	}
}
