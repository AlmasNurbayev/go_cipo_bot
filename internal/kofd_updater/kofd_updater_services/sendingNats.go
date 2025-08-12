package kofd_updater_services

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/nats-io/nats.go"
)

func SendToNats(cfg *config.Config, Log1 *slog.Logger,
	messages []models.MessagesType) error {

	op := "kofd_updater_services.SendToNats"
	log := Log1.With("op", op)
	subject := "new_transactions"

	connectionString := "nats://" + cfg.NATS_NAME + ":" + cfg.NATS_PORT
	log.Info("Connecting to NATS", "connectionString", connectionString, "topic", subject)
	nc, err := nats.Connect(connectionString,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(3),
		nats.ReconnectWait(time.Second))
	if err != nil {
		log.Error("Failed to connect to NATS", "error", err)
		return err
	}
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
				Name:     cfg.NATS_STREAM_NAME,
				Subjects: []string{subject},
				Storage:  nats.FileStorage,  // хранение на диске
				MaxBytes: 100 * 1024 * 1024, // лимит 100 MB
				Discard:  nats.DiscardOld,   // при переполнении удалять старые сообщения
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
	}

	// Публикуем сообщения в стрим
	for _, message := range messages {
		message.Sending_at = time.Now()
		data, err := json.Marshal(message)
		if err != nil {
			log.Error("Failed to marshal message", "error", err)
			return err
		}
		_, err = js.Publish(subject, data)
		if err != nil {
			log.Error("Failed to publish message", "error", err, "subject", subject)
			return err
		}
	}
	//log1.Info("All messages published successfully", "count", len(messages))
	log.Info("Message published", "subject", subject)
	defer nc.Close()
	return nil
}
