package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/kofd_updater/kofd_updater_services"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/logger"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/natsutil"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/jackc/pgx/v5"
)

func main() {
	// os.Exit пропускает defer, поэтому вызываем его только здесь - когда run
	// уже полностью развернул стек и вся очистка отработала.
	// Подробности ошибки run пишет в лог на месте
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// rollbackIfOpen откатывает транзакцию, если она ещё открыта.
// После успешного Commit pgx вернёт ErrTxClosed - это не ошибка
func rollbackIfOpen(tx pgx.Tx, log *slog.Logger) {
	// свой контекст: корневой к этому моменту может быть уже отменён по таймауту
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Error("Error rollback all db changes:", slog.String("err", err.Error()))
	}
}

func run() error {
	fmt.Println("reading config...")
	var (
		configEnv string
		firstDate string
		lastDate  string
		days      string
		bin       string
	)
	flag.StringVar(&configEnv, "configEnv", "", "Path to env-file")
	flag.StringVar(&firstDate, "firstDate", "", "Date in format YYYY-MM-DD")
	flag.StringVar(&lastDate, "lastDate", "", "Date in format YYYY-MM-DD")
	flag.StringVar(&days, "days", "", "Number of last days to update")
	flag.StringVar(&bin, "bin", "", "BIN of organization")
	flag.Parse()

	cfg := config.Mustload(configEnv)
	Log := logger.InitLogger(cfg.ENV, cfg.LOG_ERROR_PATH)
	Log.Info("=== start kofd_updater ===")

	// проверяем наличие дат
	if firstDate == "" && lastDate == "" && days == "" {
		Log.Error("not set dates - firstDate, lastDate or days")
		return errors.New("not set dates - firstDate, lastDate or days")
	}
	if days != "" {
		daysNumber, err := strconv.Atoi(days)
		if err != nil {
			Log.Error("not correct days", slog.String("err", err.Error()))
			return err
		}
		now := time.Now()
		lastDate = now.Format("2006-01-02")
		firstDate = now.AddDate(0, 0, -daysNumber).Format("2006-01-02")
	}
	Log.Info("dates", slog.String("firstDate", firstDate), slog.String("lastDate", lastDate))

	if cfg.ENV != "prod" {
		Log.Info("load config: ")
		cfgBytes, err := utils.PrintAsJSON(cfg)
		if err != nil {
			// если не удалось сериализовать конфиг, то что-то не так
			Log.Error("error: ", slog.String("err", err.Error()))
			return err
		}
		fmt.Println(string(*cfgBytes))
	}
	Log.Debug("debug message is enabled")

	// выделено в отдельную функцию, чтобы её defer-ы (закрытие пула, откат
	// транзакций) успели отработать до финального сообщения об успехе
	if err := update(cfg, Log, bin, firstDate, lastDate); err != nil {
		return err
	}

	Log.Info("=== success end kofd_updater ===")
	return nil
}

// update выполняет сам прогон: проверяет брокера, забирает операции из КОФД
// и рассылает уведомления о новых
func update(cfg *config.Config, Log *slog.Logger, bin, firstDate, lastDate string) error {
	// бюджет всего прогона: должен быть меньше интервала cron, иначе запуски
	// начнут накладываться и копить сессии в БД
	ctx, cancel := context.WithTimeout(context.Background(), cfg.POSTGRES_TIMEOUT)
	defer cancel()

	// брокер проверяем до любой работы: если публиковать новые операции будет
	// некуда, то нельзя сдвигать курсоры пользователей - уведомления пропадут
	if cfg.NATS_ENABLE {
		nc, err := natsutil.ConnectWithRetry(ctx, "nats://"+cfg.NATS_NAME+":"+cfg.NATS_PORT,
			cfg.NATS_STARTUP_TIMEOUT, Log)
		if err != nil {
			Log.Error("broker is required, aborting", slog.String("err", err.Error()))
			return err
		}
		nc.Close()
	}

	dsn := "postgres://" + cfg.POSTGRES_USER + ":" + cfg.POSTGRES_PASSWORD + "@" + cfg.POSTGRES_HOST + ":" + cfg.POSTGRES_INT_PORT + "/" + cfg.POSTGRES_DB + "?sslmode=disable"
	storage, err := storage.NewStorage(ctx, dsn, Log)
	if err != nil {
		Log.Error("not init postgres storage")
		return err
	}
	// единственное место закрытия пула - отработает на любом пути, включая панику
	defer storage.Close()

	// --- транзакция 1: загрузка операций из КОФД ---
	pgxTransaction, err := storage.Db.Begin(ctx)
	if err != nil {
		Log.Error("Not created transaction:", slog.String("err", err.Error()))
		return err
	}
	storage.Tx = &pgxTransaction
	defer rollbackIfOpen(pgxTransaction, Log)

	token, err := kofd_updater_services.GetToken(ctx, storage, Log, bin, cfg)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		return err
	}

	// загружаем транзакции за заданный период из КОФД в БД
	_, err = kofd_updater_services.GetOperationsFromApi(ctx, storage, cfg, Log, bin, token, firstDate, lastDate)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		return err
	}

	// закрываем транзакцию после записи операций
	if err := pgxTransaction.Commit(ctx); err != nil {
		// продолжать нельзя: операции не сохранены, а дальше идёт сдвиг курсоров
		// и рассылка уведомлений о том, чего в базе нет
		Log.Error("Error commit all db changes:", slog.String("err", err.Error()))
		return err
	}
	storage.Tx = nil
	Log.Info("DB changes committed")

	// --- транзакция 2: определение новых операций и рассылка ---
	pgxTransaction2, err := storage.Db.Begin(ctx)
	if err != nil {
		Log.Error("Not created transaction:", slog.String("err", err.Error()))
		return err
	}
	storage.Tx = &pgxTransaction2
	defer rollbackIfOpen(pgxTransaction2, Log)

	// определяем новые операции для каждого пользователя
	messages, err := kofd_updater_services.DetectNewOperations(ctx, storage, Log)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		return err
	}

	// отправляем операции в брокер
	if cfg.NATS_ENABLE {
		Log.Info("make new messages", slog.Int("count", len(messages)))
		if len(messages) == 0 {
			Log.Info("no new updates for users")
		} else {
			if err := kofd_updater_services.SendToNats(ctx, cfg, Log, messages); err != nil {
				// откатываем сдвиг курсоров (это сделает defer): иначе эти операции
				// больше никогда не попадут в рассылку, а пользователи их не увидят
				Log.Error("Error broker send, rolling back cursors:", slog.String("err", err.Error()))
				return err
			}
		}
	}

	// удаляем старые токены - не критично для прогона
	if err := storage.DeleteOldTokens(ctx); err != nil {
		Log.Warn("error: ", slog.String("err", err.Error()))
	}

	// опять закрываем транзакцию
	if err := pgxTransaction2.Commit(ctx); err != nil {
		Log.Error("Error commit all db changes:", slog.String("err", err.Error()))
		return err
	}
	storage.Tx = nil
	Log.Info("DB changes committed")

	return nil
}
