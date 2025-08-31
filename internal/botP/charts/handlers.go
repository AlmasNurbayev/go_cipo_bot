package charts

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
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

		if msg.Text == "график 30 дней" {
			fileBytes, sumCurrent, sumPrev, err := charts30Days(ctx, storage, log)
			if err != nil {
				log.Error("error generating chart", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ошибка генерации графика",
				})
				if err != nil {
					log.Error("error sending error", slog.String("err", err.Error()))
				}
				return
			}

			_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
				ChatID: update.Message.Chat.ID,
				Caption: "График за последние 30 дней: \n" +
					" • сейчас " + utils.FormatNumber(sumCurrent) + "\n" +
					" • год назад " + utils.FormatNumber(sumPrev) + "\n",
				Photo: &models.InputFileUpload{Filename: "chart30days.png", Data: bytes.NewReader(fileBytes)},
			})
			if err != nil {
				log.Error("error sending file", slog.String("err", err.Error()))
			}
		}

		if msg.Text == "график этот год" {
			fileBytes, sumCurrent, sumPrev, err := chartsCurrentYear(ctx, storage, log)
			if err != nil {
				log.Error("error generating chart", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ошибка генерации графика",
				})
				if err != nil {
					log.Error("error sending error", slog.String("err", err.Error()))
				}
				return
			}

			_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
				ChatID: update.Message.Chat.ID,
				Caption: "График за последние 12 месяцев: \n" +
					" • текущий год " + utils.FormatNumber(sumCurrent) + "\n" +
					" • прошлый год " + utils.FormatNumber(sumPrev) + "\n",
				Photo: &models.InputFileUpload{Filename: "chart12month.png", Data: bytes.NewReader(fileBytes)},
			})
			if err != nil {
				log.Error("error sending file", slog.String("err", err.Error()))
			}
		}

	}
}
