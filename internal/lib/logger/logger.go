package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

// InitLogger initializes a logger based on the environment.
//
// It takes a string parameter 'env' and returns a pointer to slog.Logger.
func InitLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return log
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
