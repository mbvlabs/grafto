//go:build integration
// +build integration

package handlers_test

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
