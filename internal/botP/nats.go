package botP

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/nats-io/nats.go"
)

// type NatsConsumerApp struct {
// 	Log     *slog.Logger
// 	Bot     *bot.Bot
// 	Cfg     *config.Config
// 	Ctx     context.Context
// 	Nc      *nats.Conn
// 	Js      nats.JetStreamContext
// 	Storage storageI
// }

type storageI interface {
	ListKassa(context.Context) ([]modelsI.KassaEntity, error)
	GetKassaById(context.Context, int64) (modelsI.KassaEntity, error)
}

func RunNatsConsumer(ctx context.Context, cfg *config.Config, log1 *slog.Logger, b *bot.Bot, storage storageI) error {
	op := "botP.NewNatsConsumer"
	log := log1.With("op", op)

	nc, err := nats.Connect(
		cfg.NATS_NAME+":"+cfg.NATS_PORT,
		nats.MaxReconnects(-1),            // бесконечные попытки
		nats.ReconnectWait(2*time.Second), // пауза между попытками
	)
	if err != nil {
		return err
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	// Создаём или подключаемся к consumer
	sub, err := js.PullSubscribe("new_transactions", "bot_consumer",
		nats.BindStream(cfg.NATS_STREAM_NAME),
		nats.ManualAck(),
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	log.Info("Consumer started", slog.String("stream", cfg.NATS_STREAM_NAME), slog.String("subject", "new_transactions"))

	for {
		select {
		case <-ctx.Done():
			log.Warn("Stopping consumer...")
			return nil
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(500*time.Millisecond))
			if err != nil {
				if err == nats.ErrTimeout {
					continue // просто нет сообщений
				}
				return err // обрыв или другая ошибка — выйдем, чтобы переподключиться
			}
			for _, msg := range msgs {
				var data modelsI.MessagesType
				err = json.Unmarshal(msg.Data, &data)
				if err != nil {
					log.Error("JSON unmarshal error", slog.Any("err", err))
					continue
				}
				log.Info("nats message for user_id: ", slog.Any("user_id", data.UserId))
				kassas, err := storage.ListKassa(ctx)
				if err != nil {
					log.Error("List kassa error", slog.Any("err", err))
				}

				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    data.Telegram_id,
					Text:      "Новые транзакции: " + "\n" + utils.ConvertNewOperationToMessageText(data, kassas),
					ParseMode: models.ParseModeHTML,
				})
				msg.Ack()
			}
		}
	}

}
