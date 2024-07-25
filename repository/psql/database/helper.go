package database

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// TODO: remove when controllers rework is done
func ConvertToPGTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

// TODO: remove when controllers rework is done
func ConvertFromPGTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}
