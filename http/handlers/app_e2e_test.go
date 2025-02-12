//go:build e2e
// +build e2e

package handlers_test

// func TestStoreUser(t *testing.T) {
// 	ctx := context.Background()
// 	postgres, cleanup := setupTest(t)
// 	defer cleanup()
//
// 	emailSvc := services.NewEmail()
//
// 	cacheBuilder, err := otter.NewBuilder[string, string](20)
// 	require.NoError(t, err)
//
// 	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
// 	require.NoError(t, err)
//
// 	handlers := handlers.NewHandlers(postgres, pageCacher, emailSvc)
//
// 	routes := routes.NewRoutes(handlers, nil)
// 	router, ctx := routes.SetupRoutes(ctx)
//
// 	srv := server.NewServer(ctx, router)
//
// 	go func() {
// 		srv.Start(ctx)
// 	}()
//
// 	tests := []struct {
// 		name           string
// 		payload        url.Values
// 		expectedStatus int
// 		wantErr        bool
// 	}{
// 		{
// 			name: "successful registration",
// 			payload: url.Values{
// 				"email":            {"test@example.com"},
// 				"password":         {"password123"},
// 				"confirm_password": {"password123"},
// 			},
// 			expectedStatus: http.StatusOK,
// 			wantErr:        false,
// 		},
// 		// {
// 		// 	name: "mismatched passwords",
// 		// 	payload: url.Values{
// 		// 		"email":            {"test@example.com"},
// 		// 		"password":         {"password123"},
// 		// 		"confirm_password": {"different"},
// 		// 	},
// 		// 	expectedStatus: http.StatusOK, // Returns error page but still 200
// 		// 	wantErr:        true,
// 		// },
// 		// {
// 		// 	name: "invalid email",
// 		// 	payload: url.Values{
// 		// 		"email":            {"notanemail"},
// 		// 		"password":         {"password123"},
// 		// 		"confirm_password": {"password123"},
// 		// 	},
// 		// 	expectedStatus: http.StatusOK, // Returns error page but still 200
// 		// 	wantErr:        true,
// 		// },
// 		// {
// 		// 	name: "empty password",
// 		// 	payload: url.Values{
// 		// 		"email":            {"test@example.com"},
// 		// 		"password":         {""},
// 		// 		"confirm_password": {""},
// 		// 	},
// 		// 	expectedStatus: http.StatusOK, // Returns error page but still 200
// 		// 	wantErr:        true,
// 		// },
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Create HTTP client and request
// 			jar, err := cookiejar.New(nil)
// 			require.NoError(t, err)
// 			client := &http.Client{
// 				Jar: jar,
// 			}
//
// 			// First make a GET request to /register to get the CSRF token
// 			getReq, err := http.NewRequest(
// 				http.MethodGet,
// 				"http://localhost:8080/register",
// 				nil,
// 			)
// 			require.NoError(t, err)
//
// 			getResp, err := client.Do(getReq)
// 			require.NoError(t, err)
//
// 			// Read the response body
// 			body, err := io.ReadAll(getResp.Body)
// 			require.NoError(t, err)
// 			getResp.Body.Close()
//
// 			// Extract CSRF token from the response HTML
// 			re := regexp.MustCompile(
// 				`<input type="hidden" name="gorilla.csrf.Token" value="([^"]+)"`,
// 			)
// 			matches := re.FindStringSubmatch(string(body))
// 			require.Len(t, matches, 2, "CSRF token not found in response")
// 			csrfToken := matches[1]
//
// 			tt.payload.Set("gorilla.csrf.Token", csrfToken)
// 			req, err := http.NewRequest(
// 				http.MethodPost,
// 				"http://localhost:8080/register",
// 				strings.NewReader(tt.payload.Encode()),
// 			)
// 			require.NoError(t, err)
//
// 			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
// 			// Send request
// 			resp, err := client.Do(req)
// 			require.NoError(t, err)
// 			defer resp.Body.Close()
//
// 			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
//
// 			if !tt.wantErr {
// 				// Verify user was created in database
// 				user, err := models.GetUserByEmail(
// 					context.Background(),
// 					tt.payload.Get("email"),
// 					postgres.Pool,
// 				)
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.payload.Get("email"), user.Email)
// 			}
// 		})
// 	}
// }

// func TestMain(m *testing.M) {
// 	ctx := context.Background()
// 	psql, embeddedPsql, err := psql.NewPostgresTest(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := psql.Pool.Ping(ctx); err != nil {
// 		panic(err)
// 	}
//
// 	tx, err := psql.Pool.Begin(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	defer tx.Rollback(ctx)
//
// 	slog.Info("Starting seed script...")
// 	seeder := seeds.NewSeeder(psql.Pool)
// 	user, err := seeder.PlantUser(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	slog.Info("seeded db with user", "usr", user)
//
// 	if err := tx.Commit(ctx); err != nil {
// 		panic(err)
// 	}
//
// 	emailSvc := services.NewEmail()
//
// 	cacheBuilder, err := otter.NewBuilder[string, string](20)
// 	if err != nil {
// 		panic(err)
// 	}
// 	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	handlers := handlers.NewHandlers(psql, pageCacher, emailSvc)
//
// 	routes := routes.NewRoutes(handlers, nil)
// 	router, ctx := routes.SetupRoutes(ctx)
//
// 	srv := http.NewServer(ctx, router)
//
// 	go func() {
// 		srv.Start(ctx)
// 	}()
//
// 	slog.Info("running playwright")
// 	if err := playwright.Install(); err != nil {
// 		panic(err)
// 	}
//
// 	pw, err := playwright.Run()
// 	if err != nil {
// 		panic(err)
// 	}
// 	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
// 		Headless: playwright.Bool(false),
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	context, err := browser.NewContext()
// 	if err != nil {
// 		panic(err)
// 	}
// 	page, err := context.NewPage()
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, err = page.Goto("http://localhost:8080/register")
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	if _, err := page.Locator("body > main > div > div.text-center.w-full > h1").Evaluate("el => el.textContent", "Register User"); err != nil {
// 		panic(err)
// 	}
//
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > div:nth-child(1) > label > input").
// 		Fill("testing@mbvlabs.com"); err != nil {
// 		panic(err)
// 	}
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > div:nth-child(2) > label > input").
// 		Fill("password"); err != nil {
// 		panic(err)
// 	}
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > div:nth-child(3) > label > input").
// 		Fill("password"); err != nil {
// 		panic(err)
// 	}
//
// 	time.Sleep(5 * time.Second)
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > button").
// 		Click(); err != nil {
// 		panic(err)
// 	}
//
// 	time.Sleep(5 * time.Second)
//
// 	newUser, err := models.GetUserByEmail(ctx, "testing@mbvlabs.com", psql.Pool)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	slog.Info("user successfully created", "user", newUser)
//
// 	if err := embeddedPsql.Stop(); err != nil {
// 		panic(err)
// 	}
//
// 	os.Exit(m.Run())
// }

// func TestMain(m *testing.M) {
// 	ctx := context.Background()
// 	psql, embeddedPsql, err := psql.NewPostgresTest(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := psql.Pool.Ping(ctx); err != nil {
// 		panic(err)
// 	}
//
// 	tx, err := psql.Pool.Begin(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	defer tx.Rollback(ctx)
//
// 	slog.Info("Starting seed script...")
// 	seeder := seeds.NewSeeder(psql.Pool)
// 	user, err := seeder.PlantUser(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	slog.Info("seeded db with user", "usr", user)
//
// 	if err := tx.Commit(ctx); err != nil {
// 		panic(err)
// 	}
//
// 	emailSvc := services.NewEmail()
//
// 	cacheBuilder, err := otter.NewBuilder[string, string](20)
// 	if err != nil {
// 		panic(err)
// 	}
// 	pageCacher, err := cacheBuilder.WithVariableTTL().Build()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	handlers := handlers.NewHandlers(psql, pageCacher, emailSvc)
//
// 	routes := routes.NewRoutes(handlers, nil)
// 	router, ctx := routes.SetupRoutes(ctx)
//
// 	srv := http.NewServer(ctx, router)
//
// 	go func() {
// 		srv.Start(ctx)
// 	}()
//
// 	slog.Info("running playwright")
// 	if err := playwright.Install(); err != nil {
// 		panic(err)
// 	}
//
// 	pw, err := playwright.Run()
// 	if err != nil {
// 		panic(err)
// 	}
// 	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
// 		Headless: playwright.Bool(false),
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	context, err := browser.NewContext()
// 	if err != nil {
// 		panic(err)
// 	}
// 	page, err := context.NewPage()
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, err = page.Goto("http://localhost:8080/register")
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	if _, err := page.Locator("body > main > div > div.text-center.w-full > h1").Evaluate("el => el.textContent", "Register User"); err != nil {
// 		panic(err)
// 	}
//
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > div:nth-child(1) > label > input").
// 		Fill("testing@mbvlabs.com"); err != nil {
// 		panic(err)
// 	}
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > div:nth-child(2) > label > input").
// 		Fill("password"); err != nil {
// 		panic(err)
// 	}
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > div:nth-child(3) > label > input").
// 		Fill("password"); err != nil {
// 		panic(err)
// 	}
//
// 	time.Sleep(5 * time.Second)
// 	if err := page.Locator("body > main > div > div.mt-5.w-full > form > div > button").
// 		Click(); err != nil {
// 		panic(err)
// 	}
//
// 	time.Sleep(5 * time.Second)
//
// 	newUser, err := models.GetUserByEmail(ctx, "testing@mbvlabs.com", psql.Pool)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	slog.Info("user successfully created", "user", newUser)
//
// 	if err := embeddedPsql.Stop(); err != nil {
// 		panic(err)
// 	}
//
// 	os.Exit(m.Run())
// }
