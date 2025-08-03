package kofd_updater_services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/segmentio/kafka-go"
)

func SendToKafka(cfg *config.Config, Log *slog.Logger, messages []models.MessagesType) error {

	topic := "new_transactions"
	partition := 0

	fmt.Println(cfg.KAFKA_SERVICE_NAME + ":" + cfg.KAFKA_PORT)

	conn, err := kafka.DialLeader(context.Background(),
		"tcp",
		cfg.KAFKA_SERVICE_NAME+":"+cfg.KAFKA_PORT,
		topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	for index, message := range messages {
		messages[index].Sending_at = time.Now()
		// Сериализация в JSON
		payload, err := json.Marshal(messages[index])
		if err != nil {
			Log.Error("marshal error:", slog.Any("err", err))
		}

		_, err = conn.WriteMessages(kafka.Message{
			Key:   []byte(fmt.Sprintf("%d", message.UserId)),
			Value: []byte(payload),
		},
		)
		if err != nil {
			Log.Error("failed to write messages:", slog.Any("err", err))
		}
	}

	if err := conn.Close(); err != nil {
		Log.Error("failed to close writer:", slog.Any("err", err))
	}

	return nil
}
