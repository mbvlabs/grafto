package queue

import (
	"github.com/mbv-labs/grafto/pkg/mail"
	"github.com/mbv-labs/grafto/psql/database"
	"github.com/riverqueue/river"
)

type WorkerDependencies struct {
	DB         *database.Queries
	MailClient mail.Mail
}

func SetupWorkers(deps WorkerDependencies) (*river.Workers, error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, &EmailJobWorker{
		Sender: &deps.MailClient,
	}); err != nil {
		return nil, err
	}

	return workers, nil
}
