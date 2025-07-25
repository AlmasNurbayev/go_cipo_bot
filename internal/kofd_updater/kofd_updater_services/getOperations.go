package kofd_updater_services

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/kofd_updater/api"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/guregu/null/v5"
	"golang.org/x/sync/errgroup"
)

type storageOperations interface {
	ListKassa(context.Context) ([]models.KassaEntity, error)
	ListOrganizations(context.Context) ([]models.OrganizationEntity, error)
	InsertTransactions(context.Context, []models.TransactionEntity) (int, error)
	CheckExistsTransactions(context.Context, []string) ([]string, error)
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

		// перед спамом запросов проверяем, есть ли такие транзакции в базе
		ids := make([]string, len(list.Data))
		for i, item := range list.Data {
			ids[i] = item.Id
		}
		existsIds, err := storage.CheckExistsTransactions(ctx, ids)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return 0, err
		}

		// сначала берем простые поля в цикле
		for _, item := range list.Data {

			if slices.Contains(existsIds, item.Id) {
				log.Info("transaction already exists, skip for get cheque", slog.String("id", item.Id))
				continue // если транзакция уже есть, пропускаем
			}

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
				Knumber:             kassa.Knumber,
			})
		}

		// сортируем по времени Operationdate, чтобы id отражали порядок
		slices.SortFunc(listEntity, func(a, b models.TransactionEntity) int {
			if a.Operationdate.Time.Before(b.Operationdate.Time) {
				return -1
			} else if a.Operationdate.Time.After(b.Operationdate.Time) {
				return 1
			}
			return 0
		})

		// через горутины получаем чеки
		var g errgroup.Group
		semaphore := make(chan struct{}, 10)

		for index := range listEntity {
			idx := index            // захват переменной цикла
			semaphore <- struct{}{} // занимаем слот перед запуском горутины

			g.Go(func() error {
				defer func() { <-semaphore }() // освободить слот

				cheque, err := api.KofdGetCheck(cfg, log, kassa.Knumber.String, token, listEntity[index].Ofd_id)
				if err != nil {
					log.Error("error on get check: ", slog.String("err", err.Error()))
				}
				log.Info("get and parse check from API", slog.String("id", listEntity[idx].Ofd_id))
				var sb strings.Builder
				for _, item := range cheque.Data {
					sb.WriteString(strings.TrimSpace(item.Text))
					sb.WriteString("\n")
				}
				chequeString := sb.String()
				names, err := utils.GetGoodsFromCheque(chequeString)
				if err != nil && listEntity[idx].Type_operation == 1 {
					// если продажа/возврат и не удалось получить товары из чека
					log.Error("error on get goods from cheque: ", slog.String("err", err.Error()))
					return err
				}

				// безопасно: каждый goroutine пишет только по своему индексу
				listEntity[idx].Cheque = null.StringFrom(chequeString)
				listEntity[idx].ChequeJSON = names
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return 0, err // ошибка из любой горутины
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
