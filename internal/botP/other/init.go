package other

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type storageI interface {
	ListUsers(context.Context) ([]modelsI.UserEntity, error)
}

func Init(b *bot.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) {
	// слушаем сообщения
	b.RegisterHandler(bot.HandlerTypeMessageText, "/other", bot.MatchTypeExact, initKeyboard)
	// любой регистр и любое количество символов после "финансы"
	b.RegisterHandler(bot.HandlerTypeMessageText, "other_siteParserJSONlog", bot.MatchTypeExact, otherSiteParserJSONlogHandler(log, cfg))
	b.RegisterHandler(bot.HandlerTypeMessageText, "other_sendTestMessage", bot.MatchTypeExact, otherSendTestMessageHandler(storage, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "other_sendTest_", bot.MatchTypePrefix, otherSendTestCallbackHandler(storage, log))
}

func initKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "other_siteParserJSONlog"},
			},
			{
				{Text: "other_sendTestMessage"},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите команду:",
		ReplyMarkup: kb,
	})
	if err != nil {
		fmt.Println("error sending message")
	}
}
