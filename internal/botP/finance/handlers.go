package finance

import (
	"context"
	"log/slog"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kr/pretty"
)

func financeMainHandler(log1 *slog.Logger, cfg *config.Config, settings []modelsI.SettingsEntity) bot.HandlerFunc {
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
				Text:   "запрос финансы должен быть в формате: 'финансы тек. день' или 'финансы пр. день' и т.д. или 'финансы 2024 08' или 'финансы 2024 08 21' или 'финансы 2024'",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}
		if msg.Text == "финансы произ. дата" {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text: "отправьте сообщение в формате 'финансы год месяц', например 'финансы 2024 08'\n" +
					"или 'финансы 2024' для получения итогов по году",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}
		data, err := financeOPIUService(msg.Text, settings, log)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
		pretty.Log(data)
	}
}

// func financeCallbackHandler(storage *storage.Storage, log *slog.Logger, cfg *config.Config) bot.HandlerFunc {
// 	return func(ctx context.Context, b *bot.Bot, update *models.Update) {}
// }

//log.Info("financeMainHandler")
