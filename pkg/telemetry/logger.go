package telemetry

import (
	"os"
	"time"

	"log/slog"

	"github.com/lmittmann/tint"
)

func SetupLogger() *slog.Logger {
	// create a new logger
	return slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		}),
	)

}
