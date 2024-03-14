package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
)

// main is only in place to develop emails locally
func main() {
	http.Handle("/confirm-user-email", templ.Handler(templates.ConfirmUserEmail{
		ConfirmLink:     "insert-link-here",
		UnsubscribeLink: "insert-link-here",
	}))

	fmt.Println("Listening on :4444")
	http.ListenAndServe(":4444", nil)
}
