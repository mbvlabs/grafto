package workers

import (
	"context"

	"github.com/mbvlabs/grafto/queue/jobs"
	"github.com/mbvlabs/grafto/services"
	"github.com/riverqueue/river"
)

type EmailJobWorker struct {
	email services.Email
	river.WorkerDefaults[jobs.EmailJobArgs]
}

func (w *EmailJobWorker) Work(
	ctx context.Context,
	job *river.Job[jobs.EmailJobArgs],
) error {
	return w.email.Send(
		ctx,
		services.EmailPayload{
			To:       job.Args.To,
			From:     job.Args.From,
			Subject:  job.Args.Subject,
			HtmlBody: job.Args.HtmlVersion,
			TextBody: job.Args.TextVersion,
		},
	)
}
