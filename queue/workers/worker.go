package workers

import (
	awsses "github.com/mbvlabs/grafto/pkg/aws_ses"
	"github.com/mbvlabs/grafto/pkg/telemetry"
	"github.com/mbvlabs/grafto/psql/database"
	"github.com/riverqueue/river"
)

type WorkerDependencies struct {
	DB      *database.Queries
	Emailer awsses.AwsSimpleEmailService
	Tracer  telemetry.Tracer
}

func SetupWorkers(deps WorkerDependencies) (*river.Workers, error) {
	workers := river.NewWorkers()

	if err := river.AddWorkerSafely(workers, &EmailJobWorker{
		emailer: &deps.Emailer,
	}); err != nil {
		return nil, err
	}

	return workers, nil
}
