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

func CheckUser(storage storageI, log1 *slog.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message == nil {
				next(ctx, b, update)
				return
			}

			op := "middleware.CheckUser"
			log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)),
				slog.String("user name", update.Message.From.Username))

			// ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// defer cancel()
			user, err := GetUserByTelegramId(ctx, log, storage, strconv.Itoa(int(update.Message.From.ID)))
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Error if checking user",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			if user == nil {
				log.Warn("denied: ", slog.String("err", "user not authorized"))
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "👎 User not authorized",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			if user.Role == "kaspi_manager" {
				log.Info("kaspi_manager blocked from commands", slog.String("user", user.Telegram_id))
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "ℹ️ Вам доступны только уведомления о продажах Kaspi-товаров",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			// Your middleware logic here
			next(ctx, b, update)
		}
	}
}

// GetUserByTelegramId возвращает пользователя по telegram_id или nil, если не найден
func GetUserByTelegramId(ctx context.Context, log1 *slog.Logger,
	storage storageI, telegram_id string) (*modelsI.UserEntity, error) {
	op := "middleware.GetUserByTelegramId"
	log := log1.With(slog.String("op", op))

	users, err := storage.ListUsers(ctx)
	if err != nil {
		log.Error("", slog.String("err", err.Error()))
		return nil, err
	}
	for _, user := range users {
		if user.Telegram_id == telegram_id {
			return &user, nil
		}
	}

	return nil, nil
}

// UserIsValid проверяет, есть ли пользователь в списке (обратная совместимость)
func UserIsValid(ctx context.Context, log1 *slog.Logger,
	storage storageI, telegram_id string) (bool, error) {
	user, err := GetUserByTelegramId(ctx, log1, storage, telegram_id)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

