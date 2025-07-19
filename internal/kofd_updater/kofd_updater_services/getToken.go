package kofd_updater_services

import (
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
)

func GetToken(ctx context.Context, storage storageToken, log *slog.Logger, bin string,
	cfg *config.Config) (string, error) {

	op := "kofd_updater.services.GetToken"
	log = log.With(slog.String("op", op))

	// ищем сначала токен в базе данных
	tokens, err := storage.ListActiveTokens(ctx, bin)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}
	if len(tokens) > 0 {
		log.Info("token found in DB", slog.String("token", tokens[0].Token[0:12]+"..."))
		return tokens[0].Token, nil
	} else {
		log.Info("token not found in DB, getting token from API")
		token, err := GetTokenFormApi(ctx, storage, log, bin, cfg)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return "", err
		}
		return token, nil
	}

}
