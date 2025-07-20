package summary

import (
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func summaryHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "день", CallbackData: "summary_BtnDay"},
				{Text: "неделя", CallbackData: "summary_BtnWeek"},
				{Text: "месяц", CallbackData: "summary_BtnMonth"},
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
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: cb.Message.Message.Chat.ID,
			Text:   "Вы выбрали: " + cb.Data,
		})
	}
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
