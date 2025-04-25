package workers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbvlabs/grafto/clients"
	"github.com/riverqueue/river"
)

type WorkerDependencies struct {
	DB          *pgxpool.Pool
	EmailClient clients.Email
}

func SetupWorkers(deps WorkerDependencies) (*river.Workers, error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, &EmailJobWorker{
		emailClient: deps.EmailClient,
	}); err != nil {
		return nil, err
	}

	return workers, nil
}
