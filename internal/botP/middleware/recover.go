package middleware

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// перехватывает панику в обработчиках и логирует её, после чего завершает программу
func Recover(log *slog.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("panic recovered",
						slog.Any("recover", r),
						slog.String("stack", string(debug.Stack())),
					)
					os.Exit(1)
				}
			}()
			next(ctx, b, update)
		}
	}
}
