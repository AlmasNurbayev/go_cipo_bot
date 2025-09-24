package finance

import (
	"context"
	"log/slog"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func financeMainHandler(storage *storage.Storage, log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.financeMainHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		//return
		log.Info("finance called button", slog.String("text", msg.Text))
		parts := strings.Split(msg.Text, " ")
		if len(parts) < 2 {
			log.Warn("finance called < 2 words: " + msg.Text)
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "запрос итоги должен быть в формате: 'финансы тек. день' или 'финансы пр. день' и т.д. или 'финансы 2024 08' или 'финансы 2024 08 21' или 'финансы 2024'",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}

	}
}

func financeCallbackHandler(storage *storage.Storage, log *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {}
}

//log.Info("financeMainHandler")
