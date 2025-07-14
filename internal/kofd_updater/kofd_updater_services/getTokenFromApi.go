package kofd_updater_services

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/kofd_updater/api"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

type storageToken interface {
	ListKassa(context.Context) ([]models.KassaEntity, error)
	ListOrganizations(context.Context) ([]models.OrganizationEntity, error)
	InsertToken(context.Context, string, string, int64, int64) error
	ListActiveTokens(context.Context, string) ([]models.TokenEntity, error)
}

// get token from KOFD auth by password and save it to DB
func GetTokenFormApi(ctx context.Context, storage storageToken, log *slog.Logger, bin string,
	cfg *config.Config) (string, error) {

	op := "kofd_updater.services.GetTokenFormApi"
	log = log.With(slog.String("op", op))
	token := ""

	organizationsList, err := storage.ListOrganizations(ctx)
	if err != nil {
		log.Error("", slog.String("err", err.Error()))
		return "", err
	}

	foundOrgIndex := slices.IndexFunc(organizationsList, func(org models.OrganizationEntity) bool {
		return org.Bin == bin
	})
	if foundOrgIndex == -1 {
		log.Error("Organization not found", slog.String("bin", bin))
		return "", err
	}

	password, err := utils.DecryptToken(cfg.SECRET_BYTE, organizationsList[foundOrgIndex].Hash.String)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}

	TokenRequest := api.KofdAuthenticateRequest{
		Credentials: api.Credentials{
			Iin:      organizationsList[foundOrgIndex].Bin,
			Password: password,
		},
		OrganizationXin: organizationsList[foundOrgIndex].Bin,
	}

	data, err := api.KofdGetToken(TokenRequest, cfg, log)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}
	token = data.Data.Jwt

	// передаем токен для сохранения в БД
	layout := "2006-01-02T15:04:05.999999999-07:00"
	exp, err := time.Parse(layout, data.Data.AccessTokenExpiration)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}
	nbf := time.Now().Unix()

	err = saveTokenService(storage, log, ctx,
		organizationsList[foundOrgIndex].Bin, token, exp.Unix(), nbf)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}

	return token, nil
}
