package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	slog.Info("starting")
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	slog.Info("launching")
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	slog.Info("new page")
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		// AcceptDownloads:      new(bool),
		// BaseURL:              new(string),
		// BypassCSP:            new(bool),
		// ClientCertificates:   []playwright.ClientCertificate{},
		// ColorScheme:          &"",
		// DeviceScaleFactor:    new(float64),
		// ExtraHttpHeaders: map[string]string{
		// 	"Authorization": "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjhkMjUwZDIyYTkzODVmYzQ4NDJhYTU2YWJhZjUzZmU5NDcxNmVjNTQiLCJ0eXAiOiJKV1QifQ.eyJyb2xlIjo0LCJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vbWljcm9hY3F1aXJlIiwiYXVkIjoibWljcm9hY3F1aXJlIiwiYXV0aF90aW1lIjoxNzM4ODM4MDI3LCJ1c2VyX2lkIjoiaWp4VUJobG84M1ljN2JVV0xkbjRVODlHelBWMiIsInN1YiI6ImlqeFVCaGxvODNZYzdiVVdMZG40VTg5R3pQVjIiLCJpYXQiOjE3Mzg4MzgwMjcsImV4cCI6MTczODg0MTYyNywiZW1haWwiOiJoaUBtb3J0ZW52aXN0aXNlbi5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJlbWFpbCI6WyJoaUBtb3J0ZW52aXN0aXNlbi5jb20iXX0sInNpZ25faW5fcHJvdmlkZXIiOiJwYXNzd29yZCJ9fQ.lr-DnaG5A4XKgfzjBaaMy94TllzorvLVX84G4eHC_tF9TcB84_HqQednrpt65fX3b95IJu6M4X5JFRO2aDRlnmmuyYBdsC1S8Gp3rTnEYVk7bby4Jmm8HmfDE0M6yuQyJCTQKxSiAd5PM-gFWDFL8VsUHRe1on-ZNu-w5SL5q5d5VJmYht4Ng78_DgM6WII2vXqPM4nVFnIMmvIF8TE7Ngql56YIGcZQF1VZQ7hZ43pPOMxj9NS1hMVOlmPhMCcXLcZ9yoH6Bq4pQSuPjOxbsoYAUIPRDCMnHO-ZxWPRN2Ppe3F0HqvDX5wj4-EyJE81jl0G0VzCbVOmwZBb-FhUKw",
		// },
		// ForcedColors:         &"",
		// Geolocation:          &playwright.Geolocation{},
		// HasTouch:             new(bool),
		// HttpCredentials:      &playwright.HttpCredentials{},
		// IgnoreHttpsErrors:    new(bool),
		// IsMobile:             new(bool),
		// JavaScriptEnabled:    new(bool),
		// Locale:               new(string),
		// NoViewport:           new(bool),
		// Offline:              new(bool),
		// Permissions:          []string{},
		// Proxy:                &playwright.Proxy{},
		// RecordHarContent:     &"",
		// RecordHarMode:        &"",
		// RecordHarOmitContent: new(bool),
		// RecordHarPath:        new(string),
		// RecordHarURLFilter:   nil,
		// RecordVideo:          &playwright.RecordVideo{},
		// ReducedMotion:        &"",
		// Screen:               &playwright.Size{},
		// ServiceWorkers:       &"",
		// StorageState:         &playwright.OptionalStorageState{},
		// StorageStatePath:     new(string),
		// StrictSelectors:      new(bool),
		// TimezoneId:           new(string),
		// UserAgent:            new(string),
		// Viewport:             &playwright.Size{},
	})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	slog.Info("go to")
	if _, err = page.Goto("https://app.acquire.com/browse"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	if err := page.Locator("input[type='text'].input.special-input[inputmode='email'][placeholder='richard@piedpiper.com']").Fill(
		"hi@mortenvistisen.com",
	); err != nil {
		log.Fatalf("could not fill: %v", err)
	}

	if err := page.Locator("input.input.c-topaz").Fill(
		"UmfJ62B3jttWsfU4ug7y33VBtdfxeeoutwosCoTi7Hp2qQuWZrHeph7w9Dp9Jd",
	); err != nil {
		log.Fatalf("could not fill pw: %v", err)
	}

	if err := page.Locator("#root > div > div > button.btn.btn-main.btn-action.btn-full-width.sign-in").Click(); err != nil {
		log.Fatalf("could not click: %v", err)
	}

	slog.Info("before sleep")

	time.Sleep(1 * time.Minute)
}
