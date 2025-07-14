package kofd_updater_services

import (
	"context"
	"log/slog"
)

func saveTokenService(storage storageToken, log *slog.Logger, ctx context.Context,
	bin string, token string, exp int64, nbf int64) error {

	op := "kofd_updater.services.saveTokenService"
	log = log.With(slog.String("op", op))

	err := storage.InsertToken(ctx, token, bin, exp, nbf)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return err
	}

	return nil
}
