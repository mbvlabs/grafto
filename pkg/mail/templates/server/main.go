package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
)

// main is only in place to develop emails locally
func main() {
	http.Handle("/password-reset-mail", templ.Handler(&templates.PasswordResetMail{
		ResetPasswordLink: "https://mortenvistisen.com",
		UnsubscribeLink:   "https://mortenvistisen.com",
	}))

	http.Handle("/background-job-mail", templ.Handler(&templates.BackgroundJobErrorMail{}))

	http.Handle("/user-signup-welcome-mail", templ.Handler(&templates.UserSignupWelcomeMail{
		ConfirmationLink: "https://mortenvistisen.com",
		UnsubscribeLink:  "https://mortenvistisen.com",
	}))

	http.Handle("/newsletter-welcome", templ.Handler(&templates.NewsletterWelcomeMail{
		ConfirmationLink: "https://mortenvistisen.com",
		UnsubscribeLink:  "https://mortenvistisen.com",
	}))

	fmt.Println("Listening on :4444")
	if err := http.ListenAndServe(":4444", nil); err != nil {
		panic(err)
	}
}
