// Package natsutil содержит общие для бота и апдейтера помощники работы с брокером
package natsutil

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// границы паузы между попытками подключения
const (
	retryMinWait = 2 * time.Second
	retryMaxWait = 10 * time.Second
)

// ConnectWithRetry подключается к брокеру, повторяя попытки, пока не истечёт
// timeout. Живой брокер обязателен для обоих бинарников: бот через него
// оповещает пользователей, апдейтер - публикует новые операции. Поэтому
// исчерпанный бюджет попыток это ошибка, а не повод работать дальше.
func ConnectWithRetry(ctx context.Context, url string, timeout time.Duration,
	log1 *slog.Logger, opts ...nats.Option) (*nats.Conn, error) {

	op := "natsutil.ConnectWithRetry"
	log := log1.With(slog.String("op", op))

	deadline := time.Now().Add(timeout)
	wait := retryMinWait
	attempt := 0

	for {
		attempt++
		nc, err := nats.Connect(url, opts...)
		if err == nil {
			log.Info("broker connected", slog.String("url", url), slog.Int("attempt", attempt))
			return nc, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// пауза не должна выходить за отведённый бюджет, но и бросать попытки
		// раньше срока не нужно - последнюю укорачиваем до остатка времени
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("%s: брокер %s недоступен за %s (попыток: %d): %w",
				op, url, timeout, attempt, err)
		}
		sleep := min(wait, remaining)

		log.Warn("broker unavailable, retrying",
			slog.String("url", url), slog.Int("attempt", attempt),
			slog.Duration("wait", sleep), slog.String("err", err.Error()))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}
		wait = min(wait*2, retryMaxWait)
	}
}
