package kofd_updater_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

type storageOperations2 interface {
	GetLastTransactions(context.Context, time.Time) ([]models.TransactionEntity, error)
	ListUsers(context.Context) ([]models.UserEntity, error)
	SetCursor(context.Context, int64, int64) error
}

func DetectNewOperations(ctx context.Context, storage storageOperations2,
	log *slog.Logger) ([]models.MessagesType, error) {

	op := "kofd_updater.services.DetectNewOperations"
	log = log.With(slog.String("op", op))

	before10days := time.Now().Add(-240 * time.Hour)
	operations, err := storage.GetLastTransactions(ctx, before10days)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return nil, err
	}
	log.Info("operations", slog.Int("count", len(operations)))

	users, err := storage.ListUsers(ctx)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return nil, err
	}

	var messages []models.MessagesType
	for _, user := range users {
		if user.Transaction_cursor.Int64 == 0 {
			log.Info("Чистый курсор, отправляем последнюю операцию", slog.String("user", user.Telegram_id))
			err = storage.SetCursor(ctx, operations[len(operations)-1].Id, user.Id)
			if err != nil {
				log.Error("не удалось установать курсор: ", slog.String("err", err.Error()))
				return nil, err
			}
			messages = append(messages, models.MessagesType{
				Created_at:   time.Now(),
				UserId:       user.Id,
				Transactions: []models.TransactionEntity{operations[0]},
			})
		} else {
			// в цикле ищем операции, ранее курсора
			var transactionsForMessage []models.TransactionEntity
			newCursor := int64(0)
			for _, operation := range operations {
				if operation.Id <= user.Transaction_cursor.Int64 {
					// пропускаем операции более старые, чем курсор
					continue
				}
				transactionsForMessage = append(transactionsForMessage, operation)
				// берем как курсор первую подходящую операцию, так как обратный порядок
				if newCursor == 0 {
					newCursor = operation.Id
				}
			}
			if newCursor == 0 {
				// если не нашли операций новее курсора, то пропускаем юзера
				continue
			}
			err = storage.SetCursor(ctx, newCursor, user.Id)
			if err != nil {
				log.Error("не удалось установать курсор: ", slog.String("err", err.Error()))
				return nil, err
			}
			messages = append(messages, models.MessagesType{
				Created_at:   time.Now(),
				UserId:       user.Id,
				Transactions: transactionsForMessage,
			})
			fmt.Println("Transactions in messages:", len(transactionsForMessage))
		}

	}

	//fmt.Printf("messages: %v\n", messages)

	return messages, nil
}
