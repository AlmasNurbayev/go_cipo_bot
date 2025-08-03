package botP

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumerApp struct {
	Log     *slog.Logger
	Bot     *bot.Bot
	Cfg     *config.Config
	Ctx     context.Context
	Kafka   *kafka.Reader
	Storage storageI
}

type storageI interface {
	ListKassa(context.Context) ([]modelsI.KassaEntity, error)
	GetKassaById(context.Context, int64) (modelsI.KassaEntity, error)
}

func NewKafkaReader(ctx context.Context, cfg *config.Config, log *slog.Logger, b *bot.Bot, storage storageI) (*KafkaConsumerApp, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KAFKA_SERVICE_NAME + ":" + cfg.KAFKA_PORT},
		Topic:   "new_transactions",
		GroupID: "consumer_started_from_bot",
	})

	return &KafkaConsumerApp{
		Log:     log,
		Bot:     b,
		Storage: storage,
		Cfg:     cfg,
		Ctx:     ctx,
		Kafka:   r,
	}, nil
}

func (k *KafkaConsumerApp) Run() {
	defer func() {
		err := k.Kafka.Close()
		if err != nil {
			k.Log.Error("kafka close error", slog.Any("err", err))
		}
		k.Log.Info("Kafka reader stopped")
	}() // ignore errors on closek.Kafka.Close()

	k.Log.Info("Kafka reader try to start listening: " + k.Cfg.KAFKA_SERVICE_NAME + ":" + k.Cfg.KAFKA_PORT)
	for {
		m, err := k.Kafka.FetchMessage(k.Ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("Kafka listener stopped by context")
				return
			}
			log.Printf("Kafka error: %v", err)
			continue
		}

		var data modelsI.MessagesType
		err = json.Unmarshal(m.Value, &data)
		if err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}
		k.Log.Info("Kafka message for user_id: ", slog.Any("user_id", data.UserId))
		kassas, err := k.Storage.ListKassa(k.Ctx)
		if err != nil {
			k.Log.Error("List kassa error", slog.Any("err", err))
		}

		k.Bot.SendMessage(k.Ctx, &bot.SendMessageParams{
			ChatID:    data.Telegram_id,
			Text:      "Новые транзакции: " + "\n" + utils.ConvertNewOperationToMessageText(data, kassas),
			ParseMode: models.ParseModeHTML,
		})
		if err := k.Kafka.CommitMessages(k.Ctx, m); err != nil {
			log.Printf("commit error: %v", err)
		}
	}
}
