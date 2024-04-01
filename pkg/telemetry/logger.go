package telemetry

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var Logger *slog.Logger = SetupLogger()

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
