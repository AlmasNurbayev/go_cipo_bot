package middleware

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	tele "gopkg.in/telebot.v4"
)

type storageI interface {
	ListUsers(context.Context) ([]models.UserEntity, error)
}

func CheckUser(storage storageI, log *slog.Logger, timeout time.Duration) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			op := "middleware.CheckUser"
			log = log.With(slog.String("op", op), slog.Attr(slog.Int64("id", c.Sender().ID)),
				slog.String("user name", c.Sender().Username))

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			isValid, err := UserIsValid(ctx, log, storage, strconv.Itoa(int(c.Sender().ID)))
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				return c.Send("Error if checking user")
			}
			if !isValid {
				log.Warn("denied: ", slog.String("err", "user not authorized"))
				return c.Send("👎 User not authorized")
			}
			// Your middleware logic here
			return next(c)
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
