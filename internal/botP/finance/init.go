package finance

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Init(b *bot.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) {
	// слушаем сообщения
	b.RegisterHandler(bot.HandlerTypeMessageText, "/finance", bot.MatchTypeExact, initKeyboard)
	// любой регистр и любое количество символов после "финансы"
	regSummary := regexp.MustCompile(`(?i)^финансы.*`)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regSummary, financeMainHandler(log, cfg, storage))

	//b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "finance_", bot.MatchTypePrefix, financeCallbackHandler(storage, log, cfg))

}

func initKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "финансы тек. месяц"},
				{Text: "финансы пр. месяц"},
			}, {
				{Text: "финансы тек. квартал"},
				{Text: "финансы пр. квартал"},
			},
			{
				{Text: "финансы тек. год"},
				{Text: "финансы пр. год"},
			},
			{
				{Text: "финансы произ. дата"},
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
