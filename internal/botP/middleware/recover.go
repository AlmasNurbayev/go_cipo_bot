package middleware

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// перехватывает панику в обработчиках, логирует её и отвечает пользователю.
// Процесс при этом продолжает работать - паника в хендлере одного пользователя
// не должна останавливать бота для остальных
func Recover(log *slog.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("panic recovered",
						slog.Any("recover", r),
						slog.String("stack", string(debug.Stack())),
					)
					// сообщаем пользователю, если удалось определить чат
					chatID, ok := utils.UpdateChatID(update)
					if !ok {
						return
					}
					_, err := b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: chatID,
						Text:   "Внутренняя ошибка, попробуйте позже",
					})
					if err != nil {
						log.Error("error sending message after panic", slog.String("err", err.Error()))
					}
				}
			}()
			next(ctx, b, update)
		}
	}
}
