package summary

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Init(b *bot.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) {
	// вывести ReplyMarkup - большую клавиатуру с кнопками
	b.RegisterHandler(bot.HandlerTypeMessageText, "/summary", bot.MatchTypeExact, initKeyboard)
	// услышать сообщения с большой клавиатуры
	b.RegisterHandler(bot.HandlerTypeMessageText, "итоги", bot.MatchTypePrefix, summaryHandler(storage, log, cfg))
	// услышать нажатия на inline-кнопки из сообщений
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "summary_", bot.MatchTypePrefix, summaryCallbackHandler(storage, log, cfg))

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "getCheck_", bot.MatchTypePrefix, summaryGetCheckHandler(storage, log, cfg))
}

func initKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "итоги тек. день"},
				{Text: "итоги пр. день"},
			},
			{
				{Text: "итоги тек. неделя"},
				{Text: "итоги пр. неделя"},
			},
			{
				{Text: "итоги тек. месяц"},
				{Text: "итоги пр. месяц"},
			}, {
				{Text: "итоги тек. квартал"},
				{Text: "итоги пр. квартал"},
			},
			{
				{Text: "итоги тек. год"},
				{Text: "итоги пр. год"},
			},
			{
				{Text: "итоги произ. дата"},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите период:",
		ReplyMarkup: kb,
	})
	if err != nil {
		fmt.Println("error sending message")
	}
}
