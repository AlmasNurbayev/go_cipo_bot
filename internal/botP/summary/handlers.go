package summary

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func summaryHandler(storage storageI,
	log *slog.Logger, cfg *config.Config) bot.HandlerFunc {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryButtonHandler"
		log = log.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		//return
		log.Info("summary called button", slog.String("text", msg.Text))
		parts := strings.Split(msg.Text, " ")
		if len(parts) < 3 {
			log.Warn("summary called not 3 words: " + msg.Text)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "запрос итоги должен быть в формате: 'итоги тек. день' или 'итоги пр. день' и т.д. или 'итоги 2024 08'",
			})
			return
		}

		if msg.Text == "итоги произ. месяц" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "отправьте сообщение в формате 'итоги год месяц', например 2024 08",
			})
			return
		}

		data, err := getSummaryDate(msg.Text, storage, log)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
		}
		p := message.NewPrinter(language.Russian)
		text := "<b>" + data.DateMode +
			" (" + data.StartDate.Format("02.01.2006") + " - " + data.EndDate.Format("02.01.2006") + ")</b> \n" +
			"количество чеков: " + strconv.Itoa(data.Count) + "\n" +
			"<b>чистая сумма продаж: " + p.Sprintf("%.0f", data.Sum) + "</b> \n" +
			"сумма продаж: " + p.Sprintf("%.0f", data.SumSales) + "\n" +
			" в т.ч. кеш: " + p.Sprintf("%.0f", data.SumSalesCash) + "\n" +
			"        карта: " + p.Sprintf("%.0f", data.SumSalesCard) + "\n" +
			"        смешанно: " + p.Sprintf("%.0f", data.SumSalesMixed) + "\n" +
			"        прочее: " + p.Sprintf("%.0f", data.SumSalesOther) + "\n" +
			"сумма возвратов: " + p.Sprintf("%.0f", data.SumReturns) + "\n" +
			"\n" +
			"Выемки: " + p.Sprintf("%.0f", data.SumOutputCash) + "\n" +
			"Внесения: " + p.Sprintf("%.0f", data.SumInputCash) + "\n"
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: summaryInlineKb(data.StartDate, data.EndDate),
		})
	}
}

func summaryInlineKb(data1 time.Time, data2 time.Time) *models.InlineKeyboardMarkup {
	//if strings.Contains(text, "день") || strings.Contains(text, "неделя") {
	start := data1.Format("2006-01-02")
	end := data2.Format("2006-01-02")
	fmt.Println(start, end)

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🔍 Все чеки", CallbackData: "summary_allChecks_" + start + "_" + end},
				{Text: "Аналитика", CallbackData: "summary_analytics_" + start + "_" + end},
				{Text: "Диаграмма по дням", CallbackData: "summary_chartsDay_" + start + "_" + end},
			},
		},
	}
	//}
	return nil
}

func summaryCallbackHandler(storage storageI,
	log *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryCallbackHandler"
		log = log.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username))
		if update.CallbackQuery == nil {
			return
		}
		cb := update.CallbackQuery
		log.Info("summary called callback", slog.String("data", cb.Data))
		// b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		// 	CallbackQueryID: update.CallbackQuery.ID,
		// 	ShowAlert:       false,
		// })

		if strings.Contains(cb.Data, "summary_allChecks_") {
			response, markups, err := getAllChecks(cb.Data, b, storage, log, cfg)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: cb.Message.Message.Chat.ID,
					Text:   "Ошибка получения данных",
				})
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      cb.Message.Message.Chat.ID,
				Text:        response,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: markups,
			})
		}

		if strings.Contains(cb.Data, "summary_analytics_") {
			response, markups, err := getAnalytics(cb.Data, storage, log, cfg)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: cb.Message.Message.Chat.ID,
					Text:   "Ошибка получения данных",
				})
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      cb.Message.Message.Chat.ID,
				Text:        response,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: markups,
			})
		}
		//if cb.Data == "summary_Day" {

	}
}
