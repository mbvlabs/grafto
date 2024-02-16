package queue

import (
	"context"

	"github.com/riverqueue/river"
)

const emailJobKind string = "email_job"

type EmailJobArgs struct {
	To       string      `json:"to"`
	From     string      `json:"from"`
	Subject  string      `json:"subject"`
	TmplName string      `json:"tmpl_name"`
	Payload  interface{} `json:"payload"`
}

func (EmailJobArgs) Kind() string { return emailJobKind }

type emailSender interface {
	Send(ctx context.Context, to, from, subject, tmplName string, data interface{}) error
}

type EmailJobWorker struct {
	Sender emailSender
	river.WorkerDefaults[EmailJobArgs]
}

func (w *EmailJobWorker) Work(ctx context.Context, job *river.Job[EmailJobArgs]) error {
	return w.Sender.Send(ctx, job.Args.To, job.Args.From, job.Args.Subject, job.Args.TmplName, job.Args.Payload)
}
