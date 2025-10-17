package logger

import (
	"context"
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
func InitLogger(env string, path string) *slog.Logger {
	//var log *slog.Logger

	errFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("cannot open error.log: " + err.Error())
	}
	var handler slog.Handler
	stdout := os.Stdout

	switch env {
	case envDev:
		handler = &splitHandler{
			main: tint.NewHandler(stdout, &tint.Options{
				Level: slog.LevelDebug,
			}),
			errs: slog.NewJSONHandler(errFile, &slog.HandlerOptions{
				Level: slog.LevelError,
			}),
		}
	case envProd:
		handler = &splitHandler{
			main: slog.NewJSONHandler(stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
			errs: slog.NewJSONHandler(errFile, &slog.HandlerOptions{
				Level: slog.LevelError,
			}),
		}
	}

	return slog.New(handler)
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

type splitHandler struct {
	main slog.Handler
	errs slog.Handler
}

func (h *splitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// если хоть один хендлер принимает уровень — логируем
	return h.main.Enabled(ctx, level) || h.errs.Enabled(ctx, level)
}

func (h *splitHandler) Handle(ctx context.Context, r slog.Record) error {
	// всегда логируем в основной хендлер
	if h.main.Enabled(ctx, r.Level) {
		if err := h.main.Handle(ctx, r); err != nil {
			return err
		}
	}
	// ошибки и выше — дублируем в отдельный файл
	if r.Level >= slog.LevelError && h.errs.Enabled(ctx, r.Level) {
		if err := h.errs.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (h *splitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitHandler{
		main: h.main.WithAttrs(attrs),
		errs: h.errs.WithAttrs(attrs),
	}
}

func (h *splitHandler) WithGroup(name string) slog.Handler {
	return &splitHandler{
		main: h.main.WithGroup(name),
		errs: h.errs.WithGroup(name),
	}
}
