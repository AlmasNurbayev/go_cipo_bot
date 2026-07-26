package main

import (
	"context"
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
)

func main() {
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
		return
	}
	if days != "" {
		daysNumber, err := strconv.Atoi(days)
		if err != nil {
			Log.Error("not correct days", slog.String("err", err.Error()))
			return
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
			return
		}
		fmt.Println(string(*cfgBytes))
	}
	Log.Debug("debug message is enabled")

	//fmt.Println(lastDate, firstDate, bin)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.POSTGRES_TIMEOUT)
	defer cancel()

	// брокер проверяем до любой работы: если публиковать новые операции будет
	// некуда, то нельзя сдвигать курсоры пользователей - уведомления пропадут
	if cfg.NATS_ENABLE {
		nc, err := natsutil.ConnectWithRetry(ctx, "nats://"+cfg.NATS_NAME+":"+cfg.NATS_PORT,
			cfg.NATS_STARTUP_TIMEOUT, Log)
		if err != nil {
			Log.Error("broker is required, aborting", slog.String("err", err.Error()))
			os.Exit(1)
		}
		nc.Close()
	}

	dsn := "postgres://" + cfg.POSTGRES_USER + ":" + cfg.POSTGRES_PASSWORD + "@" + cfg.POSTGRES_HOST + ":" + cfg.POSTGRES_INT_PORT + "/" + cfg.POSTGRES_DB + "?sslmode=disable"
	storage, err := storage.NewStorage(ctx, dsn, Log)
	if err != nil {
		Log.Error("not init postgres storage")
		return
	}

	pgxTransaction, err := storage.Db.Begin(ctx)
	if err != nil {
		Log.Error("Not created transaction:", slog.String("err", err.Error()))
		storage.Close()
		return
	}
	storage.Tx = &pgxTransaction

	token, err := kofd_updater_services.GetToken(ctx, storage, Log, bin, cfg)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		err = pgxTransaction.Rollback(ctx)
		if err != nil {
			Log.Error("Error rollback all db changes:", slog.String("err", err.Error()))
		}
		storage.Close()
		return
	}

	// загружаем транзакции за заданный период из КОФД в БД
	_, err = kofd_updater_services.GetOperationsFromApi(ctx, storage, cfg, Log, bin, token, firstDate, lastDate)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		err = pgxTransaction.Rollback(ctx)
		if err != nil {
			Log.Error("Error rollback all db changes:", slog.String("err", err.Error()))
		}
		storage.Close()
		return
	}

	// закрываем транзакцию после записи операций
	err = pgxTransaction.Commit(ctx)
	if err != nil {
		Log.Error("Error commit all db changes:", slog.String("err", err.Error()))
	} else {
		Log.Info("DB changes committed")
	}

	// открываем новую транзакцию
	pgxTransaction, err = storage.Db.Begin(ctx)
	if err != nil {
		Log.Error("Not created transaction:", slog.String("err", err.Error()))
		storage.Close()
		return
	}
	storage.Tx = &pgxTransaction

	// определяем новые операции для каждого пользователя
	messages, err := kofd_updater_services.DetectNewOperations(ctx, storage, Log)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		err = pgxTransaction.Rollback(ctx)
		if err != nil {
			Log.Error("Error rollback all db changes:", slog.String("err", err.Error()))
		}
		storage.Close()
		return
	}

	// отправляем операции в брокер
	if cfg.NATS_ENABLE {
		Log.Info("make new messages", slog.Int("count", len(messages)))
		if len(messages) == 0 {
			Log.Info("no new updates for users")
		} else {
			err = kofd_updater_services.SendToNats(ctx, cfg, Log, messages)
			if err != nil {
				// откатываем сдвиг курсоров: иначе эти операции больше никогда
				// не попадут в рассылку, а пользователи их не увидят
				Log.Error("Error broker send, rolling back cursors:", slog.String("err", err.Error()))
				if errRollback := pgxTransaction.Rollback(ctx); errRollback != nil {
					Log.Error("Error rollback all db changes:", slog.String("err", errRollback.Error()))
				}
				storage.Close()
				os.Exit(1)
			}
		}
	}

	// удаляем старые токены
	err = storage.DeleteOldTokens(ctx)
	if err != nil {
		Log.Warn("error: ", slog.String("err", err.Error()))
	}

	// опять закрываем транзакцию
	err = pgxTransaction.Commit(ctx)
	if err != nil {
		Log.Error("Error commit all db changes:", slog.String("err", err.Error()))
	} else {
		Log.Info("DB changes committed")
	}

	storage.Close()
	Log.Info("=== success end kofd_updater ===")

}
