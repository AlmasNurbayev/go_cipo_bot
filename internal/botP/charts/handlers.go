package charts

import (
	"context"
	"fmt"
	"log/slog"

	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func chartsHandler(storage *storage.Storage, log1 *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "charts.chartsHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		//return
		log.Info("charts called button", slog.String("text", msg.Text))

		if msg.Text == "график 30 дней прошлое" {
			bytes, err := charts30Days(ctx, storage, log)
			if err != nil {
				log.Error("error generating chart", slog.String("err", err.Error()))
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ошибка генерации графика",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			fmt.Print(string(bytes))
		}

	}
}
