package qnt

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
	// слушаем сообщения
	b.RegisterHandler(bot.HandlerTypeMessageText, "/qnt", bot.MatchTypeExact, initKeyboard)
	// любой регистр и любое количество символов после "итоги"
	b.RegisterHandler(bot.HandlerTypeMessageText, "остатки сейчас", bot.MatchTypePrefix, qntNowHandler(storage, log, cfg))
	//b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "summary_", bot.MatchTypePrefix, summaryCallbackHandler(storage, log, cfg))
	//b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "getCheck_", bot.MatchTypePrefix, summaryGetCheckHandler(storage, log, cfg))
}

func initKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "остатки сейчас"},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите вид остатков:",
		ReplyMarkup: kb,
	})
	if err != nil {
		fmt.Println("error sending message")
	}
}
