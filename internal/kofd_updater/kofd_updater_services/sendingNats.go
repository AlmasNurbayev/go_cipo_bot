package kofd_updater_services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/natsutil"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/nats-io/nats.go"
)

// окно дедупликации по Nats-Msg-Id. Дефолтные 2 минуты совпадают с интервалом
// cron апдейтера, поэтому берём с запасом - повторная публикация той же
// транзакции после сбоя не должна дать пользователю второе уведомление
const natsDuplicatesWindow = 10 * time.Minute

// SendToNats публикует по одному сообщению на транзакцию: подтверждение
// доставки в боте идёт на уровне сообщения NATS, поэтому пачка транзакций в
// одном сообщении означала бы повторную отправку уже доставленных чеков при
// любом сбое на середине пачки
func SendToNats(ctx context.Context, cfg *config.Config, Log1 *slog.Logger,
	messages []models.MessagesType) error {

	op := "kofd_updater_services.SendToNats"
	log := Log1.With("op", op)
	subject := "new_transactions"

	connectionString := "nats://" + cfg.NATS_NAME + ":" + cfg.NATS_PORT
	log.Info("Connecting to NATS", "connectionString", connectionString, "subject", subject)
	nc, err := natsutil.ConnectWithRetry(ctx, connectionString, cfg.NATS_STARTUP_TIMEOUT, Log1)
	if err != nil {
		log.Error("Failed to connect to NATS", "error", err)
		return err
	}
	defer nc.Close()
	// Создаём контекст JetStream
	js, err := nc.JetStream()
	if err != nil {
		log.Error("Failed to create JetStream", "error", err)
		return err
	}
	// Проверяем, существует ли стрим
	info, err := js.StreamInfo(cfg.NATS_STREAM_NAME)
	if err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) {
			// Стрима нет — создаём
			_, err = js.AddStream(&nats.StreamConfig{
				Name:       cfg.NATS_STREAM_NAME,
				Subjects:   []string{subject},
				Storage:    nats.FileStorage, // хранение на диске
				MaxBytes:   50 * 1024 * 1024, // лимит 50 MB
				Discard:    nats.DiscardOld,  // при переполнении удалять старые сообщения
				Duplicates: natsDuplicatesWindow,
			})
			if err != nil {
				log.Error("Failed to add stream", "error", err)
				return err
			}
		} else {
			// Другая ошибка
			log.Error("Failed to info stream", "error", err)
			return err
		}
	} else {
		log.Info("Stream already exists", "stream", info.Config.Name)
		// стрим мог быть создан раньше, без окна дедупликации - приводим к нужному
		if info.Config.Duplicates != natsDuplicatesWindow {
			streamConfig := info.Config
			streamConfig.Duplicates = natsDuplicatesWindow
			if _, err := js.UpdateStream(&streamConfig); err != nil {
				log.Error("Failed to update stream duplicates window", "error", err)
				return err
			}
			log.Info("Stream duplicates window updated", "duplicates", natsDuplicatesWindow)
		}
	}

	// Публикуем сообщения в стрим - по одному на каждую транзакцию
	published := 0
	for _, message := range messages {
		for _, transaction := range message.Transactions {
			single := message
			single.Sending_at = time.Now()
			single.Transactions = []models.TransactionEntity{transaction}

			data, err := json.Marshal(single)
			if err != nil {
				log.Error("Failed to marshal message", "error", err)
				return err
			}
			// идентификатор для дедупликации: апдейтер мог упасть после публикации,
			// но до коммита транзакции БД - тогда тот же чек будет опубликован повторно
			msgId := "tx-" + strconv.FormatInt(transaction.Id, 10) +
				"-user-" + strconv.FormatInt(single.UserId, 10)
			_, err = js.Publish(subject, data, nats.MsgId(msgId))
			if err != nil {
				log.Error("Failed to publish message", "error", err, "subject", subject, "msgId", msgId)
				return err
			}
			published++
		}
	}
	log.Info("All messages published to NATS successfully", "count", published)
	return nil
}
