package summary

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func summaryHandler(storage *storage.Storage,
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
				Text:   "запрос итоги должен быть в формате: 'итоги тек. день' или 'итоги пр. день' и т.д.",
			})
			return
		}

		data, err := dateHandler(msg.Text, b, storage, log, cfg)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
		}
		p := message.NewPrinter(language.Russian)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: "<b>" + data.DateMode +
				" (" + data.StartDate.Format("02.01.2006") + " - " + data.EndDate.Format("02.01.2006") + ")</b> \n" +
				"количество чеков: " + strconv.Itoa(data.Count) + "\n" +
				"<b>чистая сумма продаж: " + p.Sprintf("%.0f", data.Sum) + "</b> \n" +
				"сумма продаж: " + p.Sprintf("%.0f", data.SumSales) + "\n" +
				" в т.ч. кеш: " + p.Sprintf("%.0f", data.SumSalesCash) + "\n" +
				"        карта: " + p.Sprintf("%.0f", data.SumSalesCard) + "\n" +
				"        смешанно: " + p.Sprintf("%.0f", data.SumSalesMixed) + "\n" +
				"        прочее: " + p.Sprintf("%.0f", data.SumSalesOther) + "\n" +
				"сумма возвратов: " + p.Sprintf("%.0f", data.SumReturns),
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: checkInlineKb(msg.Text, data),
		})

	}
}

func dateHandler(mode string, b *bot.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) (modelsI.TypeTransactionsTotal, error) {

	op := "summary.dateHandler"
	log = log.With(slog.String("op", op))

	// Получаем границы текущего дня в локальном времени
	start, end := utils.GetDateByMode(mode)

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

func checkInlineKb(text string, data modelsI.TypeTransactionsTotal) *models.InlineKeyboardMarkup {
	if strings.Contains(text, "день") || strings.Contains(text, "неделя") {
		start := data.StartDate.Format("2006-01-02")
		end := data.EndDate.Format("2006-01-02")
		return &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "🔍 Все чеки", CallbackData: "summary_allChecks_" + start + "_" + end},
				},
			},
		}
	}
	return nil
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

func summaryCallbackHandler(storage *storage.Storage,
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
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if strings.Contains(cb.Data, "summary_allChecks_") {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    cb.Message.Message.Chat.ID,
				Text:      "<b>" + cb.Data + "</b> \n",
				ParseMode: models.ParseModeHTML,
			})
		}
		//if cb.Data == "summary_Day" {

	}
}
