package summary

import (
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

type Storage interface {
	ListKassa(context.Context) ([]models.KassaEntity, error)
	ListOrganizations(context.Context) ([]models.OrganizationEntity, error)
	ListUsers(context.Context) ([]models.UserEntity, error)
}

func ListKassaService(ctx context.Context, log *slog.Logger, storage Storage) ([]models.KassaEntity, error) {
	op := "summary.services.ListKassaService"
	log = log.With(slog.String("op", op))

	kassa, err := storage.ListKassa(ctx)
	if err != nil {
		log.Error("", slog.String("err", err.Error()))
		return kassa, err
	}

	return kassa, nil
}

func ListOrganizationsService(ctx context.Context, log *slog.Logger, storage Storage) ([]models.OrganizationEntity, error) {
	op := "summary.services.ListOrganizationsService"
	log = log.With(slog.String("op", op))

	organizations, err := storage.ListOrganizations(ctx)
	if err != nil {
		log.Error("", slog.String("err", err.Error()))
		return organizations, err
	}

	return organizations, nil
}
