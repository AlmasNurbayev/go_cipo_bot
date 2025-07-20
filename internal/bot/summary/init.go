package summary

import (
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
)

func Init(b *bot.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/summary", bot.MatchTypeExact, summaryHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "summary_", bot.MatchTypePrefix, summaryCallbackHandler(storage, log, cfg))
}
