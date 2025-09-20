package summary

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func summaryHandler(storage storageI,
	log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryButtonHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		//return
		log.Info("summary called button", slog.String("text", msg.Text))
		parts := strings.Split(msg.Text, " ")
		if len(parts) < 2 {
			log.Warn("summary called < 2 words: " + msg.Text)
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "запрос итоги должен быть в формате: 'итоги тек. день' или 'итоги пр. день' и т.д. или 'итоги 2024 08' или 'итоги 2024 08 21' или 'итоги 2024'",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}

		if msg.Text == "итоги произ. дата" {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text: "отправьте сообщение в формате 'итоги год месяц', например 'итоги 2024 08'\n" +
					"или 'итоги 2024 08 21' для получения итогов по конкретному дню",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}

		data, err := getSummaryDate(msg.Text, storage, log)
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
		p := message.NewPrinter(language.Russian)
		text := "<b>" + data.DateMode +
			" (" + data.StartDate.Format("02.01.2006") + " - " + data.EndDate.Format("02.01.2006") + ")</b> \n" +
			"количество чеков: " + strconv.Itoa(data.Count) + "\n" +
			"<b>чистая сумма продаж: " + p.Sprintf("%.0f", data.Sum) + "</b> \n" +
			"сумма продаж: " + p.Sprintf("%.0f", data.SumSales) + "\n" +
			"сумма возвратов: " + p.Sprintf("%.0f", data.SumReturns) + "\n" +
			" в т.ч. кеш: " + p.Sprintf("%.0f", data.SumSalesCash) + "\n" +
			"        карта: " + p.Sprintf("%.0f", data.SumSalesCard) + "\n" +
			"        смешанно: " + p.Sprintf("%.0f", data.SumSalesMixed) + "\n" +
			"        прочее: " + p.Sprintf("%.0f", data.SumSalesOther) + "\n" +
			"сумма возвратов: " + p.Sprintf("%.0f", data.SumReturns) + "\n" +
			"\n"

		// если есть больше 1 кассы, то выводим информацию по ним
		if len(data.KassaTotal) > 1 {
			text += "по кассам:\n"
		}

		for _, kassa := range data.KassaTotal {
			// если нет чеков по кассе или одна касса, то пропускаем
			if kassa.Count == 0 || len(data.KassaTotal) == 1 {
				continue
			}
			text += "<b>" + kassa.NameKassa + "</b> (" + kassa.NameOrganization + ") " +
				"кол-во чеков: " + strconv.Itoa(kassa.Count) + "\n" +
				"чистая сумма продаж: " + p.Sprintf("%.0f", kassa.Sum) + "\n" +
				"сумма продаж: " + p.Sprintf("%.0f", kassa.SumSales) + "\n" +
				"сумма возвратов: " + p.Sprintf("%.0f", kassa.SumReturns) + "\n"
		}
		text +=
			"\nВыемки: " + p.Sprintf("%.0f", data.SumOutputCash) + "\n" +
				"Внесения: " + p.Sprintf("%.0f", data.SumInputCash) + "\n\n" +
				"<b>Наличие денег в кассах: </b>" + "\n"

		for _, kassa := range data.KassaTotal {
			if kassa.CashAmount != 0 {
				text += kassa.NameKassa +
					" остаток: " + p.Sprintf("%.0f", kassa.CashAmount) + "\n"
			}
		}

		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: summaryInlineKb(data.StartDate, data.EndDate),
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}
	}
}

func summaryInlineKb(data1 time.Time, data2 time.Time) *models.InlineKeyboardMarkup {
	//if strings.Contains(text, "день") || strings.Contains(text, "неделя") {
	start := data1.Format("2006-01-02")
	end := data2.Format("2006-01-02")
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🔍 Все чеки", CallbackData: "summary_allChecks_" + start + "_" + end},
				{Text: "Аналитика", CallbackData: "summary_analytics_" + start + "_" + end},
				{Text: "Диаграмма по дням", CallbackData: "summary_chartsDay_" + start + "_" + end},
			},
		},
	}
}

func summaryCallbackHandler(storage storageI,
	log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryCallbackHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username))
		if update.CallbackQuery == nil {
			return
		}
		cb := update.CallbackQuery
		log.Info("called callback", slog.String("data", cb.Data))
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}

		if strings.Contains(cb.Data, "summary_allChecks_") {
			response, markups, err := getAllChecksService(cb.Data, storage, log, cfg)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: cb.Message.Message.Chat.ID,
					Text:   "Ошибка получения данных",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
			}
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      cb.Message.Message.Chat.ID,
				Text:        response,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: markups,
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}

		if strings.Contains(cb.Data, "summary_analytics_") {
			err := utils.SendAction(cb.Message.Message.Chat.ID, "typing", b)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
			}
			response, markups, err := getAnalyticsService(cb.Data, storage, log, cfg)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: cb.Message.Message.Chat.ID,
					Text:   "Ошибка получения данных",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
			}
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      cb.Message.Message.Chat.ID,
				Text:        response,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: markups,
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
		//if cb.Data == "summary_Day" {

	}
}

func summaryGetCheckHandler(storage storageI,
	log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryGetCheckHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username), slog.String("test", "test"))
		if update.CallbackQuery == nil {
			return
		}
		cb := update.CallbackQuery
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}
		err = utils.SendAction(cb.Message.Message.Chat.ID, "upload_photo", b)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
		}
		inputMedia, stringResponce, err := getOneCheckService(cb.Data, storage, log, cfg)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: cb.Message.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}

		// Если фото есть, то отправляем МедиаГруппой, иначе просто текстом
		if len(*inputMedia) == 0 {
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    cb.Message.Message.Chat.ID,
				Text:      stringResponce,
				ParseMode: models.ParseModeHTML,
				ReplyMarkup: &models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{
								Text:         "Полный текст чека",
								CallbackData: "getFullTextCheck_" + strings.Split(cb.Data, "_")[1],
							},
						},
					},
				},
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		} else {
			// если есть фото, отправляем медиа группой
			_, err = b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
				ChatID: cb.Message.Message.Chat.ID,
				Media:  *inputMedia,
			})
			if err != nil {
				// если не получилось отправить медиа группой, то отправляем текстом
				log.Error("error sending media group", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    cb.Message.Message.Chat.ID,
					Text:      stringResponce,
					ParseMode: models.ParseModeHTML,
					ReplyMarkup: &models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{
									Text:         "Полный текст чека",
									CallbackData: "getFullTextCheck_" + strings.Split(cb.Data, "_")[1],
								},
							},
						},
					},
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
			}
			// так как нельзя отправить кнопку к МедиаГруппе, то отправляем отдельным сообщением
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: cb.Message.Message.Chat.ID,
				Text:   "Полный текст чека",
				ReplyMarkup: &models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{
								Text:         "Открыть",
								CallbackData: "getFullTextCheck_" + strings.Split(cb.Data, "_")[1],
							},
						},
					},
				},
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
	}
}

func summaryFullTextCheckHandler(storage storageI,
	log1 *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryFullCheckHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username), slog.String("test", "test"))
		if update.CallbackQuery == nil {
			return
		}
		cb := update.CallbackQuery
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}

		response, err := getFullTextCheckService(cb.Data, storage, log)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: cb.Message.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    cb.Message.Message.Chat.ID,
			Text:      response,
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}

	}
}
