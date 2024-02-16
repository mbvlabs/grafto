package queue

import (
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/riverqueue/river"
)

type WorkerDependencies struct {
	Db         *database.Queries
	MailClient mail.Mail
}

func SetupWorkers(deps WorkerDependencies) (*river.Workers, error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, &EmailJobWorker{
		Sender: &deps.MailClient,
	}); err != nil {
		return nil, err
	}

	if err := river.AddWorkerSafely(workers, &RemoveUnverifiedUsersJobWorker{
		Storage: deps.Db,
	}); err != nil {
		return nil, err
	}

	return workers, nil
}
