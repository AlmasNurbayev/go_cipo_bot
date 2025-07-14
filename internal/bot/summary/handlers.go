package summary

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	tele "gopkg.in/telebot.v4"
)

func CurentDay(b *tele.Bot, c tele.Context, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) error {

	log = log.With(slog.String("op", "summary.handlers.CurentDay"),
		slog.String("user name", c.Sender().Username),
		slog.Attr(slog.Int64("id", c.Sender().ID)))
	log.Info("incoming request")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.BOT_TIMEOUT)
	defer cancel()

	// kassa, err := ListKassaService(ctx, log, storage)
	// if err != nil {
	// 	log.Error("error: ", slog.String("err", err.Error()))
	// 	return c.Send("Ошибка получения данных")
	// }
	organizations, err := ListOrganizationsService(ctx, log, storage)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return c.Send("Ошибка получения данных")
	}
	str, err := json.Marshal(organizations)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return c.Send("Ошибка получения данных")
	}
	pass, err := utils.DecryptToken(cfg.SECRET_BYTE, organizations[0].Hash.String)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return c.Send("Ошибка получения данных")
	}

	return c.Send("Сводка за день " + string(str) + " пароль " + pass)
}
