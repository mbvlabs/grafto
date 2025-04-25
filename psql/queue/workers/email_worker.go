package workers

import (
	"context"

	"github.com/mbvlabs/grafto/clients"
	"github.com/mbvlabs/grafto/psql/queue/jobs"
	"github.com/riverqueue/river"
)

type EmailJobWorker struct {
	emailClient clients.Email
	river.WorkerDefaults[jobs.EmailJobArgs]
}

func (w *EmailJobWorker) Work(
	ctx context.Context,
	job *river.Job[jobs.EmailJobArgs],
) error {
	return w.emailClient.Send(
		ctx,
		clients.EmailPayload{
			To:       job.Args.To,
			From:     job.Args.From,
			Subject:  job.Args.Subject,
			HtmlBody: job.Args.HtmlVersion,
			TextBody: job.Args.TextVersion,
		},
		clients.Unsubscribe{},
	)
}
