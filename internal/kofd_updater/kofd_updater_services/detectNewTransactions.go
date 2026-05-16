package kofd_updater_services

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

type storageOperations2 interface {
	GetLastTransactions(context.Context, time.Time) ([]models.TransactionEntity, error)
	ListUsers(context.Context) ([]models.UserEntity, error)
	SetCursor(context.Context, int64, int64) error
}

func DetectNewOperations(ctx context.Context, storage storageOperations2,
	log1 *slog.Logger) ([]models.MessagesType, error) {

	op := "kofd_updater.services.DetectNewOperations"
	log := log1.With(slog.String("op", op))
	var messages []models.MessagesType

	before10days := time.Now().Add(-240 * time.Hour)
	operations, err := storage.GetLastTransactions(ctx, before10days)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return messages, err
	}
	if len(operations) == 0 {
		return messages, nil
	}
	log.Info("get last operations", slog.Int("count", len(operations)))

	users, err := storage.ListUsers(ctx)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return messages, err
	}

	for _, user := range users {
		log.Info("cursor for user ", slog.String("user", user.Telegram_id), slog.Int("id", int(user.Transaction_cursor.Int64)))
		if user.Transaction_cursor.Int64 == 0 {
			log.Info("Чистый курсор, отправляем последнюю операцию", slog.String("user", user.Telegram_id))
			err = storage.SetCursor(ctx, operations[0].Id, user.Id)
			if err != nil {
				log.Error("не удалось установать курсор: ", slog.String("err", err.Error()))
				return messages, err
			}
			// для kaspi_manager фильтруем: только продажи/возвраты с kaspi товарами
			if user.Role == "kaspi_manager" {
				if hasKaspiInSale(operations[0]) {
					messages = append(messages, models.MessagesType{
						Created_at:   time.Now(),
						UserId:       user.Id,
						Telegram_id:  user.Telegram_id,
						Transactions: []models.TransactionEntity{operations[0]},
					})
				}
			} else {
				messages = append(messages, models.MessagesType{
					Created_at:   time.Now(),
					UserId:       user.Id,
					Telegram_id:  user.Telegram_id,
					Transactions: []models.TransactionEntity{operations[0]},
				})
			}
		} else {
			// в цикле ищем операции, ранее курсора
			var transactionsForMessage []models.TransactionEntity
			newCursor := int64(0)
			for _, operation := range operations {
				if operation.Id <= user.Transaction_cursor.Int64 {
					// пропускаем операции более старые, чем курсор
					continue
				}
				// для kaspi_manager: включаем только продажи/возвраты с kaspi товарами
				if user.Role == "kaspi_manager" {
					if hasKaspiInSale(operation) {
						transactionsForMessage = append(transactionsForMessage, operation)
					}
				} else {
					transactionsForMessage = append(transactionsForMessage, operation)
				}
				// курсор сдвигаем всегда, независимо от роли
				if newCursor == 0 {
					newCursor = operation.Id
				}
			}
			if newCursor == 0 {
				// если не нашли операций новее курсора, то пропускаем юзера
				continue
			}
			// сдвигаем курсор для всех пользователей
			err = storage.SetCursor(ctx, newCursor, user.Id)
			if err != nil {
				log.Error("не удалось установать курсор: ", slog.String("err", err.Error()))
				return messages, err
			}
			// добавляем сообщение только если есть подходящие транзакции
			if len(transactionsForMessage) > 0 {

				messages = append(messages, models.MessagesType{
					Created_at:   time.Now(),
					UserId:       user.Id,
					Telegram_id:  user.Telegram_id,
					Transactions: transactionsForMessage,
				})
				log.Info("Transactions in message:", slog.Int("count", len(transactionsForMessage)),
					slog.String("role", user.Role))
			}
		}

	}

	//fmt.Printf("messages: %v\n", messages)

	return messages, nil
}

// hasKaspiInSale проверяет, является ли транзакция продажей/возвратом
// и содержит ли хотя бы один товар с признаком kaspi_in_sale
func hasKaspiInSale(tx models.TransactionEntity) bool {
	if tx.Type_operation != 1 {
		return false // не продажа/возврат
	}
	return slices.ContainsFunc(tx.ChequeJSON, func(item models.ChequeJSONElement) bool {
		return item.KaspiInSale
	})
}
