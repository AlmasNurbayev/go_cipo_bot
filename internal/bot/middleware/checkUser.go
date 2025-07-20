package middleware

import (
	"context"
	"log/slog"
	"strconv"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type storageI interface {
	ListUsers(context.Context) ([]modelsI.UserEntity, error)
}

func CheckUser(storage storageI, log *slog.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message == nil {
				next(ctx, b, update)
				return
			}

			op := "middleware.CheckUser"
			log = log.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)),
				slog.String("user name", update.Message.From.Username))

			// ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// defer cancel()
			isValid, err := UserIsValid(ctx, log, storage, strconv.Itoa(int(update.Message.From.ID)))
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Error if checking user",
				})
				return
			}
			if !isValid {
				log.Warn("denied: ", slog.String("err", "user not authorized"))
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "👎 User not authorized",
				})
				return
			}
			// Your middleware logic here
			next(ctx, b, update)
		}
	}
}

func UserIsValid(ctx context.Context, log *slog.Logger,
	storage storageI, telegram_id string) (bool, error) {
	op := "summary.services.ListUsersService"
	log = log.With(slog.String("op", op))

	users, err := storage.ListUsers(ctx)
	if err != nil {
		log.Error("", slog.String("err", err.Error()))
		return false, err
	}
	for _, user := range users {
		if user.Telegram_id == telegram_id {
			return true, nil
		}
	}

	return false, nil
}
