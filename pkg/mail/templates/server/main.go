package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
	"github.com/riverqueue/river/rivertype"
	"github.com/golang-module/carbon/v2"
)

// main is only in place to develop emails locally
func main() {
	http.Handle("/password-reset-mail", templ.Handler(&templates.PasswordResetMail{
		ResetPasswordLink: "https://mortenvistisen.com",
	}))

	http.Handle("/background-job-mail", templ.Handler(&templates.BackgroundJobErrorMail{
		JobID:           0,
		AttemptedAt:     time.Now(),
		Kind:            "",
		MetaData:        "",
		Err:             errors.New("could not finish job"),
		AttemptedErrors: []rivertype.AttemptError{
			{
				At:      time.Now(),
				Attempt: 2,
				Error:   "bad connection",
				Trace:   "trace trace",
			},
			{
				At:      carbon.Now().SubDay().StdTime(),
				Attempt: 1,
				Error:   "bad token",
				Trace:   "trace",
			},
		},
	}))

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
