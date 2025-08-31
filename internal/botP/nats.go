package botP

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/nats-io/nats.go"
)

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

	stream, err := js.StreamInfo(cfg.NATS_STREAM_NAME)
	if err != nil {
		// если стрима нет, создаём
		_, err = js.AddStream(&nats.StreamConfig{
			Name:      cfg.NATS_STREAM_NAME,
			Subjects:  []string{"new_transactions"},
			Storage:   nats.FileStorage, // или MemoryStorage
			Retention: nats.LimitsPolicy,
			MaxBytes:  -1,
		})
		if err != nil {
			log.Error("failed to create stream", slog.Any("err", err))
		}
		log.Info("stream created", slog.String("stream", cfg.NATS_STREAM_NAME))
	} else {
		log.Info("stream already exists", slog.String("stream", stream.Config.Name))
	}

	// Создаём или подключаемся к consumer
	sub, err := js.PullSubscribe("new_transactions", "bot_consumer",
		nats.BindStream(cfg.NATS_STREAM_NAME),
		nats.ManualAck(),
	)
	if err != nil {
		return err
	}
	defer func() {
		err := sub.Unsubscribe()
		if err != nil {
			log.Error("Error unsubscribing from NATS", slog.Any("err", err))
		}
	}()

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
				// формируем клавиатуру с кнопками чеков
				for _, tr := range data.Transactions {
					var keyboardButtons []models.InlineKeyboardButton
					keyboardButtons = append(keyboardButtons, models.InlineKeyboardButton{
						Text:         "чек №" + strconv.Itoa(int(tr.Id)) + " / " + utils.FormatNumber(tr.Sum_operation.Float64) + "₸",
						CallbackData: "getCheck_" + strconv.Itoa(int(tr.Id)),
					})
					markups := models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							keyboardButtons,
						},
					}
					// формируем текст сообщения и отправляем
					text := "Новые транзакции: " + "\n" + utils.ConvertNewOperationToMessageText(tr, kassas)
					isDisabled := true
					_, err = b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID:    data.Telegram_id,
						Text:      text,
						ParseMode: models.ParseModeHTML,
						LinkPreviewOptions: &models.LinkPreviewOptions{
							IsDisabled: &isDisabled,
						},
						ReplyMarkup: &markups,
					})
					if err != nil {
						log.Error("Send message error", slog.Any("err", err))
					}
				}

				err = msg.Ack()
				if err != nil {
					log.Error("Message ack error", slog.Any("err", err))
				}
			}
		}
	}

}
