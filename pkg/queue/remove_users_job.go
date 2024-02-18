package queue

import (
	"context"
	"time"

	"github.com/mbv-labs/grafto/repository/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/riverqueue/river"
)

const removeUnverifiedUsersJobKind string = "remove_unverified_users"

type RemoveUnverifiedUsersJobArgs struct{}

func (RemoveUnverifiedUsersJobArgs) Kind() string { return removeUnverifiedUsersJobKind }

// func (RemoveUnverifiedUsersJobArgs) InsertOpts() river.InsertOpts {
// 	return river.InsertOpts{
// 		UniqueOpts: river.UniqueOpts{
// 			ByArgs:   true,
// 			ByPeriod: 24 * time.Hour,
// 		},
// 	}
// }

type storage interface {
	RemoveUnverifiedUsers(ctx context.Context, twoWeeksAgo pgtype.Timestamptz) error
}

type RemoveUnverifiedUsersJobWorker struct {
	Storage storage
	river.WorkerDefaults[RemoveUnverifiedUsersJobArgs]
}

func (w *RemoveUnverifiedUsersJobWorker) Work(ctx context.Context, job *river.Job[RemoveUnverifiedUsersJobArgs]) error {
	return w.Storage.RemoveUnverifiedUsers(ctx, database.ConvertToPGTimestamptz(time.Now().AddDate(0, 0, -14)))
}
