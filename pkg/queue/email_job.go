package queue

import (
	"context"

	"github.com/riverqueue/river"
)

const emailJobKind string = "email_job"

type EmailJobArgs struct {
	To          string `json:"to"`
	From        string `json:"from"`
	Subject     string `json:"subject"`
	TextVersion string `json:"text_version"`
	HtmlVersion string `json:"html_version"`
}

func (EmailJobArgs) Kind() string { return emailJobKind }

type emailSender interface {
	Send(
		ctx context.Context,
		to,
		from,
		subject,
		textVersion,
		htmlVersion string,
	) error
}

type EmailJobWorker struct {
	Sender emailSender
	river.WorkerDefaults[EmailJobArgs]
}

func (w *EmailJobWorker) Work(ctx context.Context, job *river.Job[EmailJobArgs]) error {
	return w.Sender.Send(
		ctx,
		job.Args.To,
		job.Args.From,
		job.Args.Subject,
		job.Args.TextVersion,
		job.Args.HtmlVersion,
	)
}
