package jobs

import "context"

const emailJobKind string = "email_job"

type EmailJobArgs struct {
	To          string `json:"to"`
	From        string `json:"from"`
	Subject     string `json:"subject"`
	TextVersion string `json:"text_version"`
	HtmlVersion string `json:"html_version"`
}

func (EmailJobArgs) Kind() string { return emailJobKind }

type EmailSender interface {
	Send(
		ctx context.Context,
		to,
		from,
		subject,
		textVersion,
		htmlVersion string,
	) error
}
