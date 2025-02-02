package workers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbvlabs/grafto/services"
	"github.com/riverqueue/river"
)

type WorkerDependencies struct {
	DB       *pgxpool.Pool
	EmailSvc services.Email
}

func SetupWorkers(deps WorkerDependencies) (*river.Workers, error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, &EmailJobWorker{
		email: deps.EmailSvc,
	}); err != nil {
		return nil, err
	}

	return workers, nil
}
