package kofd_updater_services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/kofd_updater/api"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/guregu/null/v5"
)

type storageOperations interface {
	ListKassa(context.Context) ([]models.KassaEntity, error)
	ListOrganizations(context.Context) ([]models.OrganizationEntity, error)
	InsertTransactions(context.Context, []models.TransactionEntity) (int, error)
}

func GetOperationsFromApi(ctx context.Context, storage storageOperations, cfg *config.Config, log *slog.Logger,
	BIN string, token string, firstDate string, lastDate string) (int, error) {

	op := "kofd_updater.services.GetOperationsService"
	log = log.With(slog.String("op", op))

	// ctx, cancel := context.WithTimeout(context.Background(), cfg.POSTGRES_TIMEOUT)
	// defer cancel()

	//var transactions []models.TransactionEntity
	requestKassaList, err := storage.ListKassa(ctx)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return 0, err
	}

	newCount := 0 // счетчик вставленных строк

	for _, kassa := range requestKassaList {
		if !kassa.Is_active {
			continue
		}
		list, err := api.KofdGetOperations(cfg, log, kassa.Knumber.String, token, firstDate, lastDate)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return 0, err
		}
		listEntity := []models.TransactionEntity{}
		for _, item := range list.Data {

			check, err := api.KofdGetCheck(cfg, log, kassa.Knumber.String, token, item.Id)
			if err != nil {
				log.Error("error on get check: ", slog.String("err", err.Error()))
			}

			var sb strings.Builder
			//log.Info("get check", slog.String("id", item.Id), slog.Int("count", len(check.Data)))
			for _, item := range check.Data {
				sb.WriteString(strings.TrimSpace(item.Text))
				sb.WriteString("\n")
			}
			checkString := sb.String()
			names, err := utils.GetGoodsFromCheque(checkString)
			if err != nil && item.Type_operation == 1 {
				// если продажа/возврат и не удалось получить товары из чека
				log.Error("error on get goods from cheque: ", slog.String("err", err.Error()))
				return 0, err
			}
			fmt.Println("names", names)

			//log.Info("checkString", slog.String("checkString", checkString))

			listEntity = append(listEntity, models.TransactionEntity{
				Ofd_id:              item.Id,
				Ofd_name:            null.StringFrom("KOFD"),
				Offlinefiscalnumber: item.OfflineFiscalNumber,
				Onlinefiscalnumber:  item.OnlineFiscalNumber,
				Systemdate:          null.NewTime(item.SystemDate, !item.SystemDate.IsZero()),
				Operationdate:       null.NewTime(item.OperationDate, !item.OperationDate.IsZero()),
				Type_operation:      item.Type_operation,
				Subtype:             null.IntFrom(int64(item.SubType)),
				Sum_operation:       null.FloatFrom(item.Sum_operation),
				Availablesum:        null.FloatFrom(item.AvailableSum),
				Paymenttypes:        &item.PaymentTypes,
				Shift:               null.IntFrom(int64(item.Shift)),
				Organization_id:     kassa.Organization_id,
				Kassa_id:            kassa.Id,
				Cheque:              null.StringFrom(checkString),
				Knumber:             kassa.Knumber,
			})
		}
		log.Info("get transactions from api", slog.Int("count", len(listEntity)), slog.String("kassa", kassa.Knumber.String))
		count, err := storage.InsertTransactions(ctx, listEntity)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return 0, err
		}
		newCount += count

	}
	return newCount, nil
}
