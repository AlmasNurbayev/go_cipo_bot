package qnt

import (
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func qntNowHandler(storage *storage.Storage, log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "qnt.qntNowHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		//return
		log.Info("qnt now called button", slog.String("text", msg.Text))

		text, err := qntNowService(log, cfg)
		if err != nil {
			log.Error("error qntNowService", slog.String("err", err.Error()))
		}

		const telegramMaxLen = 4096
		runes := []rune(text) // чтобы не порезать UTF-8 посередине
		for start := 0; start < len(runes); start += telegramMaxLen {
			end := start + telegramMaxLen
			if end > len(runes) {
				end = len(runes)
			}
			part := string(runes[start:end])

			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      part,
				ParseMode: models.ParseModeHTML,
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}

		// if msg.Text == "график 30 дней" {
		// 	fileBytes, sumCurrent, sumPrev, err := charts30Days(ctx, storage, log)
		// 	if err != nil {
		// 		log.Error("error generating chart", slog.String("err", err.Error()))
		// 		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		// 			ChatID: update.Message.Chat.ID,
		// 			Text:   "Ошибка генерации графика",
		// 		})
		// 		if err != nil {
		// 			log.Error("error sending error", slog.String("err", err.Error()))
		// 		}
		// 		return
		// 	}

		// 	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		// 		ChatID: update.Message.Chat.ID,
		// 		Caption: "График за последние 30 дней: \n" +
		// 			" • сейчас " + utils.FormatNumber(sumCurrent) + "\n" +
		// 			" • год назад " + utils.FormatNumber(sumPrev) + "\n",
		// 		Photo: &models.InputFileUpload{Filename: "chart30days.png", Data: bytes.NewReader(fileBytes)},
		// 	})
		// 	if err != nil {
		// 		log.Error("error sending file", slog.String("err", err.Error()))
		// 	}
		// }

		// if msg.Text == "график этот год" {
		// 	fileBytes, sumCurrent, sumPrev, err := chartsCurrentYear(ctx, storage, log)
		// 	if err != nil {
		// 		log.Error("error generating chart", slog.String("err", err.Error()))
		// 		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		// 			ChatID: update.Message.Chat.ID,
		// 			Text:   "Ошибка генерации графика",
		// 		})
		// 		if err != nil {
		// 			log.Error("error sending error", slog.String("err", err.Error()))
		// 		}
		// 		return
		// 	}

		// 	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		// 		ChatID: update.Message.Chat.ID,
		// 		Caption: "График этот год': \n" +
		// 			" • текущий год " + utils.FormatNumber(sumCurrent) + "\n" +
		// 			" • прошлый год " + utils.FormatNumber(sumPrev) + "\n",
		// 		Photo: &models.InputFileUpload{Filename: "chart12month.png", Data: bytes.NewReader(fileBytes)},
		// 	})
		// 	if err != nil {
		// 		log.Error("error sending file", slog.String("err", err.Error()))
		// 	}
		// }

	}
}
