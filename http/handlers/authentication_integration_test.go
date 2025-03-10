//go:build integration
// +build integration

package handlers_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/http/handlers"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/models/seeds"
	"github.com/mbvlabs/grafto/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreAuthenticatedSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	testHandlers := setupTestHandlers(t, postgres)
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
				fmt.Sprintf(
					"%s/%s",
					config.Cfg.GetFullDomain(),
					"login",
				),
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

func TestStorePasswordReset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	testHandlers := setupTestHandlers(t, postgres)
	router, ctx := setupTestRouter(ctx, testHandlers)

	seeder := seeds.NewSeeder(postgres.Pool)
	validUser, err := seeder.PlantUser(
		ctx,
		seeds.WithUserEmailVerifiedAt(time.Now()),
		seeds.WithUserEmail("jonsnow@gmail.com"),
	)
	assert.NoError(t, err)

	tests := []struct {
		name              string
		user              models.UserEntity
		payload           url.Values
		expectedToSucceed bool
	}{
		{
			name: "should send password reset",
			user: validUser,
			payload: url.Values{
				"email": {validUser.Email},
			},
			expectedToSucceed: true,
		},
		{
			name: "should not send password reset",
			user: models.UserEntity{},
			payload: url.Values{
				"email": {"doesnotexist@gmail.com"},
			},
			expectedToSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				fmt.Sprintf(
					"%s/%s",
					config.Cfg.GetFullDomain(),
					"forgot-password",
				),
				strings.NewReader(tt.payload.Encode()),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			c := router.NewContext(req, rec)

			var sentHtml string

			if tt.expectedToSucceed {
				emailSvc.On(
					"Send",
					ctx,
					mock.MatchedBy(func(payload services.EmailPayload) bool {
						correctEmail := payload.To == tt.user.Email
						correctSubject := payload.Subject == "Action Required | Password reset requested"

						sentHtml = payload.HtmlBody

						if correctEmail && correctSubject {
							return true
						}

						return false
					}),
				).Return(nil)
			}
			if !tt.expectedToSucceed {
				if ok := emailSvc.AssertNotCalled(
					t,
					"Send",
					ctx,
					services.EmailPayload{},
				); !ok {
					assert.FailNow(
						t,
						"Send method was called when it should not be",
					)
				}
			}

			err := testHandlers.Authentication.StorePasswordReset(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			if tt.expectedToSucceed {
				doc, err := goquery.NewDocumentFromReader(
					bytes.NewBuffer([]byte(sentHtml)),
				)
				assert.NoError(t, err)

				href, ok := doc.Find("a.link").First().Attr("href")
				assert.True(t, ok)

				assert.NotEmpty(t, href, "reset password link was empty")

				token := strings.Split(href, "?token=")[1]

				resetPwTkn, err := models.GetToken(ctx, postgres.Pool, token)
				assert.NoError(t, err)

				assert.True(t, resetPwTkn.IsValid())
			}
		})
	}
}

func TestStoreResetPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	postgres, cleanup, stopEmbedded := setupTestDB(ctx, t)
	defer cleanup()
	defer stopEmbedded()

	testHandlers := setupTestHandlers(t, postgres)
	router, ctx := setupTestRouter(ctx, testHandlers)

	seeder := seeds.NewSeeder(postgres.Pool)
	validUser, err := seeder.PlantUser(
		ctx,
		seeds.WithUserEmailVerifiedAt(time.Now()),
		seeds.WithUserEmail("jonsnow@gmail.com"),
	)
	assert.NoError(t, err)

	validToken, err := seeder.PlantToken(
		ctx,
		seeds.WithTokenExpiration(time.Now().Add(1*time.Hour)),
		seeds.WithTokenMeta(models.MetaInformation{
			Resource:   models.ResourceUser,
			ResourceID: validUser.ID,
			Scope:      models.ScopeResetPassword,
		}),
	)
	assert.NoError(t, err)

	validUserWithInvalidTkn, err := seeder.PlantUser(
		ctx,
		seeds.WithUserEmailVerifiedAt(time.Now()),
		seeds.WithUserEmail("aryastark@gmail.com"),
	)
	assert.NoError(t, err)
	expiredToken, err := seeder.PlantToken(
		ctx,
		seeds.WithTokenExpiration(time.Now().Add(-1*time.Hour)),
		seeds.WithTokenMeta(models.MetaInformation{
			Resource:   models.ResourceUser,
			ResourceID: validUserWithInvalidTkn.ID,
			Scope:      models.ScopeResetPassword,
		}),
	)
	assert.NoError(t, err)

	invalidScopedToken, err := seeder.PlantToken(
		ctx,
		seeds.WithTokenExpiration(time.Now().Add(1*time.Hour)),
		seeds.WithTokenMeta(models.MetaInformation{
			Resource:   models.ResourceUser,
			ResourceID: validUserWithInvalidTkn.ID,
			Scope:      models.ScopeEmailVerification,
		}),
	)
	assert.NoError(t, err)

	tests := []struct {
		name              string
		token             models.Token
		payload           url.Values
		expectedToSucceed bool
	}{
		{
			name:  "should reset password successfully",
			token: validToken,
			payload: url.Values{
				"password":         {"reset_password"},
				"confirm_password": {"reset_password"},
				"token":            {validToken.Hash},
			},
			expectedToSucceed: true,
		},
		{
			name:  "should not reset password bc invalid scope",
			token: invalidScopedToken,
			payload: url.Values{
				"password":         {"reset_password"},
				"confirm_password": {"reset_password"},
				"token":            {invalidScopedToken.Hash},
			},
			expectedToSucceed: false,
		},
		{
			name:  "should not reset password bc expired token",
			token: expiredToken,
			payload: url.Values{
				"password":         {"reset_password"},
				"confirm_password": {"reset_password"},
				"token":            {expiredToken.Hash},
			},
			expectedToSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				fmt.Sprintf(
					"%s/%s",
					config.Cfg.GetFullDomain(),
					"reset-password",
				),
				strings.NewReader(tt.payload.Encode()),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rec := httptest.NewRecorder()
			c := router.NewContext(req, rec)

			err := testHandlers.Authentication.StoreResetPassword(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			if tt.expectedToSucceed {
				user, err := models.GetUser(
					ctx,
					tt.token.Meta.ResourceID,
					postgres.Pool,
				)
				assert.NoError(t, err)

				assert.NoError(t, user.ValidatePassword("reset_password"))

				_, err = models.GetToken(ctx, postgres.Pool, tt.token.Hash)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			}

			if !tt.expectedToSucceed {
				user, err := models.GetUser(
					ctx,
					tt.token.Meta.ResourceID,
					postgres.Pool,
				)
				assert.NoError(t, err)

				assert.NoError(t, user.ValidatePassword("password"))

				_, err = models.GetToken(ctx, postgres.Pool, tt.token.Hash)
				assert.NoError(t, err)
			}
		})
	}
}
