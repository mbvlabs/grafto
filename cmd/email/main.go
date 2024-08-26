package main

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gosimple/slug"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/views/emails"
)

func getTextVersionSlugs() ([]string, error) {
	files, err := emails.TextTemplates.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var textVersions []string
	for _, f := range files {
		removedSuffix, _ := strings.CutSuffix(f.Name(), ".txt")
		splitted := strings.Split(removedSuffix, "_")

		var finished string
		for _, el := range splitted {
			finished = slug.Make(finished + " " + el)
		}

		textVersions = append(textVersions, finished)
	}

	return textVersions, nil
}

func getHtmlVersionSlugs() ([]string, error) {
	files, err := emails.HtmlTemplates.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var htmlVersions []string
	for _, f := range files {
		removedSuffix, _ := strings.CutSuffix(f.Name(), "_templ.go")
		splitted := strings.Split(removedSuffix, "_")

		var finished string
		for _, el := range splitted {
			finished = slug.Make(finished + " " + el)
		}

		htmlVersions = append(htmlVersions, finished)
	}

	return htmlVersions, nil
}

func main() {
	textEmails, err := getTextVersionSlugs()
	if err != nil {
		panic(err)
	}

	htmlEmails, err := getHtmlVersionSlugs()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	index := emailsIndex(textEmails, htmlEmails)
	e.GET("/", func(c echo.Context) error {
		return index.Render(c.Request().Context(), c.Response())
	})

	passwordReset := emails.PasswordReset{
		ResetPasswordLink: "wvSwI8Yq02o9cmJ6zVSTkP44lXGJZjmMF8v10vxAhrrV6UyzRr59ogUzdo3VKP7y",
	}
	userSignupWelcome := emails.UserSignupWelcome{
		ConfirmationLink: "wvSwI8Yq02o9cmJ6zVSTkP44lXGJZjmMF8v10vxAhrrV6UyzRr59ogUzdo3VKP7y",
	}

	textGroup := e.Group("/text-emails")
	textGroup.GET("/password-reset", func(c echo.Context) error {
		tmpl, err := template.ParseFS(emails.TextTemplates, "password_reset.txt")
		if err != nil {
			slog.Error("error", "e", err)
			return c.HTML(http.StatusInternalServerError, "could not parse template")
		}

		return tmpl.Execute(c.Response().Writer, passwordReset)
	})
	textGroup.GET("/user-signup-welcome", func(c echo.Context) error {
		return userSignupWelcome.Render(c.Request().Context(), c.Response())
	})

	htmlGroup := e.Group("/html-emails")
	htmlGroup.GET("/password-reset", func(c echo.Context) error {
		return passwordReset.Render(c.Request().Context(), c.Response())
	})
	htmlGroup.GET("/user-signup-welcome", func(c echo.Context) error {
		return userSignupWelcome.Render(c.Request().Context(), c.Response())
	})

	// http.Handle("/password-reset-mail", templ.Handler(&emails.PasswordReset{
	// 	ResetPasswordLink: "https://mortenvistisen.com",
	// }))
	// http.Handle("/user-signup-welcome-mail", templ.Handler(&emails.UserSignupWelcomeMail{
	// 	ConfirmationLink: "https://mortenvistisen.com",
	// }))

	slog.Info("starting the password server on port: 4444")
	log.Fatal(e.Start(":4444"))
}
