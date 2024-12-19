package workers

import (
	"context"

	emails "github.com/mbvlabs/grafto/pkg/email_client"
	"github.com/mbvlabs/grafto/queue/jobs"
	"github.com/riverqueue/river"
)

type EmailJobWorker struct {
	emailer emails.EmailClient
	river.WorkerDefaults[jobs.EmailJobArgs]
}

func (w *EmailJobWorker) Work(ctx context.Context, job *river.Job[jobs.EmailJobArgs]) error {
	return w.emailer.Send(
		ctx,
		job.Args.To,
		job.Args.From,
		job.Args.Subject,
		job.Args.HtmlVersion,
		job.Args.TextVersion,
	)
}
