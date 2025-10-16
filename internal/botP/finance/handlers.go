package finance

import (
	"bytes"
	"context"
	"log/slog"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func financeMainHandler(log1 *slog.Logger, cfg *config.Config, storage *storage.Storage) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.financeMainHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		cb := update.Message
		err := utils.SendAction(ctx, cb.Chat.ID, "typing", b)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
		}

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

		var data []byte
		var text string

		// если не диаграмма, то получаем данные и текст
		if !(strings.Contains(msg.Text, "диаграмма")) {
			var err2 error
			// формируем данные и картинку таблицы
			data, text, err2 = financeOPIUService(ctx, log, storage, msg.Text, cfg.GOOGLE_API_KEY)
			if err2 != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ошибка получения данных",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
			}
			//return
		} else {
			var err2 error
			data, text, err2 = financeChartService(ctx, log, storage, msg.Text, cfg.GOOGLE_API_KEY)
			if err2 != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ошибка получения данных",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				//return
			}
		}

		// отправляем картинку таблицы
		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:    update.Message.Chat.ID,
			Caption:   text,
			ParseMode: models.ParseModeHTML,
			Photo:     &models.InputFileUpload{Filename: "chart30days.png", Data: bytes.NewReader(data)},
		})
		if err != nil {
			log.Error("error sending file", slog.String("err", err.Error()))
		}
	}
}
