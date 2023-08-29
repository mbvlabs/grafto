package telemetry

import (
	"os"
	"time"

	"github.com/lmittmann/tint"
	"log/slog"
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
