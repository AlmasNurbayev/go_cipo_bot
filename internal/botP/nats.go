package botP

import (
	"context"
	"encoding/json"
	"errors"
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

const (
	natsSubject       = "new_transactions"
	natsConsumerName  = "bot_consumer"
	natsMaxAckPending = 64
	natsFetchBatch    = 5
	// потолок времени на обработку всей пачки Fetch: если бот умрёт посередине,
	// брокер отдаст неподтверждённые сообщения заново только после этого срока.
	// Должен быть заметно больше, чем natsFetchBatch * telegramSendTimeout -
	// иначе при медленном Telegram брокер начнёт дублировать сообщения, которые
	// мы ещё обрабатываем
	natsAckWait = 5 * time.Minute
	// пауза перед повторной доставкой сообщения, которое не удалось отправить
	// из-за временного сбоя Telegram
	natsNakDelay = 30 * time.Second
	// с какого момента читать стрим, если консьюмера ещё нет: окно должно
	// покрывать простой бота при деплое, но не всю историю стрима
	natsConsumerLookback = 15 * time.Minute
	// таймаут одной отправки в Telegram; у http-клиента бота свой лимит в минуту,
	// но пачка из natsFetchBatch сообщений должна укладываться в natsAckWait
	telegramSendTimeout = 30 * time.Second
	// конфигурация стрима должна совпадать с той, что задаёт публикатор
	// (kofd_updater_services.SendToNats) - иначе кто первый создал, того и настройки
	natsStreamMaxBytes   = 50 * 1024 * 1024
	natsDuplicatesWindow = 10 * time.Minute
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
// того, что подписка была успешно создана.
//
// Контракт подтверждения: Ack отправляется только после того, как сообщение
// действительно ушло в Telegram, либо после ошибки, которую повтор не исправит.
// Всё остальное - Nak с задержкой, повторную доставку делает брокер. Это важно,
// потому что курсор рассылки (users.transaction_cursor) сдвигается апдейтером
// независимо от доставки: подтверждённое, но не отправленное сообщение теряется
// навсегда
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
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return false, err
		}
		// если стрима нет, создаём
		_, err = js.AddStream(&nats.StreamConfig{
			Name:       cfg.NATS_STREAM_NAME,
			Subjects:   []string{natsSubject},
			Storage:    nats.FileStorage, // или MemoryStorage
			Retention:  nats.LimitsPolicy,
			MaxBytes:   natsStreamMaxBytes,
			Discard:    nats.DiscardOld,
			Duplicates: natsDuplicatesWindow,
		})
		if err != nil {
			log.Error("failed to create stream", slog.Any("err", err))
			return false, err
		}
		log.Info("stream created", slog.String("stream", cfg.NATS_STREAM_NAME))
	} else {
		log.Info("stream already exists", slog.String("stream", stream.Config.Name))
	}

	// консьюмер настраиваем явно, до подписки: у созданного по умолчанию AckWait
	// всего 30 секунд, а нам нужен запас на всю пачку Fetch
	if err := ensureConsumer(js, cfg.NATS_STREAM_NAME, log); err != nil {
		return false, err
	}

	// подключаемся к уже настроенному консьюмеру - опции, меняющие его конфиг,
	// здесь передавать нельзя, иначе получим рассинхрон с существующим durable
	sub, err := js.PullSubscribe(natsSubject, "",
		nats.Bind(cfg.NATS_STREAM_NAME, natsConsumerName),
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

	log.Info("Consumer started", slog.String("stream", cfg.NATS_STREAM_NAME), slog.String("subject", natsSubject))

	for {
		select {
		case <-ctx.Done():
			log.Warn("Stopping consumer...")
			return true, nil
		default:
			msgs, err := sub.Fetch(natsFetchBatch, nats.MaxWait(500*time.Millisecond))
			if err != nil {
				if err == nats.ErrTimeout {
					continue // просто нет сообщений
				}
				return true, err // обрыв или другая ошибка — выйдем, чтобы переподключиться
			}
			for i, msg := range msgs {
				if ctx.Err() != nil {
					// приложение останавливается - возвращаем брокеру всё, что
					// не успели обработать, иначе после рестарта ждать бы их
					// пришлось до истечения natsAckWait
					nakRest(log, msgs[i:])
					log.Warn("Stopping consumer...")
					return true, nil
				}
				if err := handleNatsMessage(ctx, log, b, storage, msg); err != nil {
					log.Error("delivery failed, message returned to broker",
						slog.Any("err", err), slog.Duration("delay", nakDelay(err)))
					if nakErr := msg.NakWithDelay(nakDelay(err)); nakErr != nil {
						log.Error("Message nak error", slog.Any("err", nakErr))
					}
					// остальные сообщения обрабатываем: один залипший
					// пользователь не должен блокировать очередь остальным
					continue
				}
				if err := msg.Ack(); err != nil {
					// не критично: сообщение придёт повторно, пользователь
					// увидит дубль - это лучше, чем потерянный чек
					log.Error("Message ack error", slog.Any("err", err))
				}
			}
		}
	}

}

// handleNatsMessage отправляет пользователю все транзакции из сообщения.
// Возвращённая ошибка означает "подтверждать нельзя, нужна повторная доставка";
// nil - сообщение доставлено либо провалилось необратимо и держать его в очереди
// бессмысленно
func handleNatsMessage(ctx context.Context, log *slog.Logger, b *bot.Bot,
	storage storageI, msg *nats.Msg) error {

	var data modelsI.MessagesType
	err := json.Unmarshal(msg.Data, &data)
	if err != nil {
		// повторная доставка битого JSON ничего не изменит - подтверждаем,
		// иначе сообщение будет крутиться в очереди вечно
		log.Error("JSON unmarshal error", slog.Any("err", err))
		return nil
	}
	log.Info("nats message for user_id: ", slog.Any("user_id", data.UserId))

	kassas, err := storage.ListKassa(ctx)
	if err != nil {
		// без списка касс текст сообщения будет неполным, лучше повторить позже
		log.Error("List kassa error", slog.Any("err", err))
		return err
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

		sendCtx, cancel := context.WithTimeout(ctx, telegramSendTimeout)
		_, err = b.SendMessage(sendCtx, &bot.SendMessageParams{
			ChatID:    data.Telegram_id,
			Text:      text,
			ParseMode: models.ParseModeHTML,
			LinkPreviewOptions: &models.LinkPreviewOptions{
				IsDisabled: &isDisabled,
			},
			ReplyMarkup: &markups,
		})
		cancel()
		if err == nil {
			continue
		}

		log.Error("Send message error", slog.Any("err", err),
			slog.Any("ChatID", data.Telegram_id), slog.Any("user_id", data.UserId))
		if isPermanentTelegramError(err) {
			// бот заблокирован, чат не найден или текст не принят Telegram -
			// повторять бессмысленно, идём дальше
			continue
		}
		return err
	}
	return nil
}

// isPermanentTelegramError отделяет ошибки, которые не исправит повторная
// отправка. Список известных ошибок в go-telegram/bot неполный: сетевые сбои и
// коды 5xx приходят безымянной ошибкой, поэтому всё неопознанное считаем
// временным - потерять чек хуже, чем прислать его повторно.
// ErrorUnauthorized (отозванный токен) тоже временная: это авария конфигурации,
// после починки токена сообщения должны дойти
func isPermanentTelegramError(err error) bool {
	if errors.Is(err, bot.ErrorForbidden) || // пользователь заблокировал бота
		errors.Is(err, bot.ErrorBadRequest) || // чат не найден, битая разметка
		errors.Is(err, bot.ErrorNotFound) {
		return true
	}
	return bot.IsMigrateError(err) // чат стал супергруппой, id больше не действителен
}

// nakDelay - пауза перед повторной доставкой. Обычно фиксированная, но при 429
// Telegram сам говорит, сколько ждать, и раньше срока пробовать смысла нет
func nakDelay(err error) time.Duration {
	var tooManyRequests *bot.TooManyRequestsError
	if errors.As(err, &tooManyRequests) {
		retryAfter := time.Duration(tooManyRequests.RetryAfter) * time.Second
		if retryAfter > natsNakDelay {
			return retryAfter
		}
	}
	return natsNakDelay
}

// nakRest возвращает брокеру необработанный хвост пачки
func nakRest(log *slog.Logger, msgs []*nats.Msg) {
	for _, msg := range msgs {
		if err := msg.Nak(); err != nil {
			log.Error("Message nak error", slog.Any("err", err))
		}
	}
}

// ensureConsumer создаёт durable-консьюмер или приводит существующий к нужным
// настройкам. Консьюмер, созданный неявно при подписке, имеет AckWait 30 секунд -
// этого не хватает на пачку сообщений при медленном Telegram, и брокер начнёт
// отдавать их повторно ещё до того, как мы ответим
func ensureConsumer(js nats.JetStreamContext, stream string, log *slog.Logger) error {
	info, err := js.ConsumerInfo(stream, natsConsumerName)
	if err != nil {
		if !errors.Is(err, nats.ErrConsumerNotFound) {
			return err
		}
		// консьюмера нет - создаём. Читать стрим с самого начала нельзя:
		// в нём лежит история за 50 Мб, и все пользователи разом получат
		// уведомления о давно прошедших чеках. Берём небольшое окно назад,
		// чтобы не потерять чеки, накопившиеся за время простоя бота
		startTime := time.Now().Add(-natsConsumerLookback)
		_, err = js.AddConsumer(stream, &nats.ConsumerConfig{
			Durable:       natsConsumerName,
			FilterSubject: natsSubject,
			DeliverPolicy: nats.DeliverByStartTimePolicy,
			OptStartTime:  &startTime,
			AckPolicy:     nats.AckExplicitPolicy,
			AckWait:       natsAckWait,
			MaxDeliver:    -1, // недоставленный чек не выбрасываем никогда
			MaxAckPending: natsMaxAckPending,
		})
		if err != nil {
			log.Error("failed to create consumer", slog.Any("err", err))
			return err
		}
		log.Info("consumer created", slog.String("consumer", natsConsumerName),
			slog.Time("start_time", startTime))
		return nil
	}

	if info.Config.AckWait == natsAckWait &&
		info.Config.MaxDeliver == -1 &&
		info.Config.MaxAckPending == natsMaxAckPending {
		return nil
	}
	// у существующего консьюмера правим только параметры подтверждения:
	// политику доставки и точку старта сервер менять не даёт
	wanted := info.Config
	wanted.AckWait = natsAckWait
	wanted.MaxDeliver = -1
	wanted.MaxAckPending = natsMaxAckPending
	if _, err := js.UpdateConsumer(stream, &wanted); err != nil {
		log.Error("failed to update consumer", slog.Any("err", err))
		return err
	}
	log.Info("consumer config updated", slog.String("consumer", natsConsumerName),
		slog.Duration("ack_wait", natsAckWait))
	return nil
}
