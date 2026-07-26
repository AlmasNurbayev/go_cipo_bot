package botP

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/natsutil"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/nats-io/nats.go"
)

type storageI interface {
	ListKassa(context.Context) ([]modelsI.KassaEntity, error)
	GetKassaById(context.Context, int64) (modelsI.KassaEntity, error)
	ListUsers(context.Context) ([]modelsI.UserEntity, error)
}

// границы паузы между попытками переподключения к брокеру в рабочем режиме
const (
	natsRetryMinWait = 2 * time.Second
	natsRetryMaxWait = 60 * time.Second
)

// RunNatsConsumer состоит из двух фаз.
//
// Старт: брокер обязателен - оповещение пользователей о новых чеках одна из
// основных задач бота. Попытки подключения повторяются в пределах
// NATS_STARTUP_TIMEOUT, после чего возвращается ошибка и приложение падает.
//
// Работа: потеря брокера бота не останавливает, консьюмер переподключается
// бесконечно. Об аварии один раз оповещаются админы; флаг сбрасывается после
// восстановления, чтобы о следующем обрыве тоже сообщили.
func RunNatsConsumer(ctx context.Context, cfg *config.Config, log1 *slog.Logger, b *bot.Bot, storage storageI) error {
	op := "botP.RunNatsConsumer"
	log := log1.With("op", op)

	// фаза старта - без брокера работать не начинаем
	nc, err := natsutil.ConnectWithRetry(ctx, cfg.NATS_NAME+":"+cfg.NATS_PORT,
		cfg.NATS_STARTUP_TIMEOUT, log1,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil // остановка приложения во время ожидания брокера
		}
		return err
	}
	nc.Close() // проверочное соединение больше не нужно, дальше подключается консьюмер

	// фаза работы
	wait := natsRetryMinWait
	adminsNotified := false
	for {
		subscribed, err := runNatsConsumerOnce(ctx, cfg, log1, b, storage)
		if ctx.Err() != nil {
			// штатная остановка вместе с приложением
			return nil
		}
		if subscribed {
			// связь была - сбрасываем задержку и право снова оповестить админов
			wait = natsRetryMinWait
			adminsNotified = false
		}
		if err == nil {
			continue
		}
		log.Error("nats consumer failed, retrying",
			slog.String("err", err.Error()), slog.Duration("wait", wait))

		if !adminsNotified {
			notifyAdmins(ctx, log, b, storage, err)
			adminsNotified = true
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
		// увеличиваем паузу до потолка, чтобы не долбить недоступный брокер
		wait = min(wait*2, natsRetryMaxWait)
	}
}

// notifyAdmins один раз за аварию сообщает всем админам, что брокер потерян и
// уведомления о новых чеках временно не приходят
func notifyAdmins(ctx context.Context, log *slog.Logger, b *bot.Bot, storage storageI, cause error) {
	users, err := storage.ListUsers(ctx)
	if err != nil {
		log.Error("cannot list users to notify about broker", slog.String("err", err.Error()))
		return
	}
	text := "⚠️ Брокер сообщений недоступен, уведомления о новых чеках временно не приходят.\n" +
		"Причина: " + cause.Error()

	sent := 0
	for _, user := range users {
		if user.Role != "admin" {
			continue
		}
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.Telegram_id,
			Text:   text,
		})
		if err != nil {
			log.Error("error notifying admin about broker",
				slog.String("telegram_id", user.Telegram_id), slog.String("err", err.Error()))
			continue
		}
		sent++
	}
	log.Warn("admins notified about broker outage", slog.Int("count", sent))
}

// runNatsConsumerOnce - одна попытка: подключение, подписка и чтение сообщений
// до первой ошибки или отмены контекста. Первым значением возвращает признак
// того, что подписка была успешно создана
func runNatsConsumerOnce(ctx context.Context, cfg *config.Config, log1 *slog.Logger, b *bot.Bot, storage storageI) (bool, error) {
	op := "botP.runNatsConsumerOnce"
	log := log1.With("op", op)

	nc, err := nats.Connect(
		cfg.NATS_NAME+":"+cfg.NATS_PORT,
		nats.MaxReconnects(-1),            // бесконечные попытки
		nats.ReconnectWait(2*time.Second), // пауза между попытками
	)
	if err != nil {
		return false, err
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return false, err
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
		return false, err
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
			return true, nil
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(500*time.Millisecond))
			if err != nil {
				if err == nats.ErrTimeout {
					continue // просто нет сообщений
				}
				return true, err // обрыв или другая ошибка — выйдем, чтобы переподключиться
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
						log.Error("Send message error", slog.Any("err", err), slog.Any("ChatID", data.Telegram_id), slog.Any("user_id", data.UserId))
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
