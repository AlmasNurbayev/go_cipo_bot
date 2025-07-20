package summary

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func summaryHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "текущий день", CallbackData: "summary_day_current"},
				{Text: "текущая неделя", CallbackData: "summary_week_current"},
				{Text: "текущий месяц", CallbackData: "summary_month_current"},
			},
			{
				{Text: "прошлый день", CallbackData: "summary_day_previous"},
				{Text: "прошлая неделя", CallbackData: "summary_week_previous"},
				{Text: "прошлый месяц", CallbackData: "summary_month_previous"},
			},
		},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите период:",
		ReplyMarkup: kb,
	})
}

func summaryCallbackHandler(storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryButtonHandler"
		log = log.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username))
		if update.CallbackQuery == nil {
			return
		}
		cb := update.CallbackQuery
		log.Info("summary called callback", slog.String("data", cb.Data))
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		//if cb.Data == "summary_Day" {
		log.Info("summary current day")
		data, err := dateHandler(cb.Data, b, storage, log, cfg)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: cb.Message.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
		}
		p := message.NewPrinter(language.Russian)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: cb.Message.Message.Chat.ID,
			Text: "режим: " + data.DateMode + "\n" +
				"нач. дата: " + data.StartDate.String() + "\n" +
				"кон. дата: " + data.EndDate.String() + "\n" +
				"количество чеков: " + strconv.Itoa(data.Count) + "\n" +
				"<b>чистая сумма продаж: " + p.Sprintf("%.0f", data.Sum) + "</b> \n" +
				"сумма продаж: " + p.Sprintf("%.0f", data.SumSales) + "\n" +
				"сумма возвратов: " + p.Sprintf("%.0f", data.SumReturns),
			ParseMode: models.ParseModeHTML,
		})
		//}

		// b.SendMessage(ctx, &bot.SendMessageParams{
		// 	ChatID: cb.Message.Message.Chat.ID,
		// 	Text:   "Вы выбрали: " + cb.Data,
		// })
	}
}

func dateHandler(mode string, b *bot.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) (modelsI.TypeTransactionsTotal, error) {

	op := "summary.dateHandler"
	log = log.With(slog.String("op", op))

	// Получаем границы текущего дня в локальном времени
	start, end := getDateByMode(mode)

	var result = modelsI.TypeTransactionsTotal{
		StartDate: start,
		EndDate:   end,
		DateMode:  mode,
	}

	data, err := storage.ListTransactionsByDate(context.Background(), start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	log.Info("transactions count", slog.Int("count", len(data)))

	result = utils.ConvertTransToTotal(result, data)

	return result, nil
}

// func CurentDay(b *tele.Bot, c tele.Context, storage *storage.Storage,
// 	log *slog.Logger, cfg *config.Config) error {

// 	log = log.With(slog.String("op", "summary.handlers.CurentDay"),
// 		slog.String("user name", c.Sender().Username),
// 		slog.Attr(slog.Int64("id", c.Sender().ID)))
// 	log.Info("incoming request")

// 	ctx, cancel := context.WithTimeout(context.Background(), cfg.BOT_TIMEOUT)
// 	defer cancel()

// 	// kassa, err := ListKassaService(ctx, log, storage)
// 	// if err != nil {
// 	// 	log.Error("error: ", slog.String("err", err.Error()))
// 	// 	return c.Send("Ошибка получения данных")
// 	// }
// 	organizations, err := ListOrganizationsService(ctx, log, storage)
// 	if err != nil {
// 		log.Error("error: ", slog.String("err", err.Error()))
// 		return c.Send("Ошибка получения данных")
// 	}
// 	str, err := json.Marshal(organizations)
// 	if err != nil {
// 		log.Error("error: ", slog.String("err", err.Error()))
// 		return c.Send("Ошибка получения данных")
// 	}
// 	pass, err := utils.DecryptToken(cfg.SECRET_BYTE, organizations[0].Hash.String)
// 	if err != nil {
// 		log.Error("error: ", slog.String("err", err.Error()))
// 		return c.Send("Ошибка получения данных")
// 	}

// 	return c.Send("Сводка за день " + string(str) + " пароль " + pass)
// }

func getDateByMode(mode string) (time.Time, time.Time) {

	parts := strings.Split(mode, "_")

	start, end := time.Now(), time.Now()

	if parts[1] == "day" && parts[2] == "current" {
		now := time.Now()
		loc := now.Location()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end = start.Add(24 * time.Hour).Add(-time.Nanosecond)
		//return start, end
	} else if parts[1] == "day" && parts[2] == "previous" {
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour).Add(-time.Nanosecond)
		//return start, end
	} else if mode == "month" {
		//return start, end
	} else if mode == "year" {
		//return start, end
	}
	return start, end
}
