package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/kofd_updater/kofd_updater_services"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/logger"
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
	Log := logger.InitLogger(cfg.ENV)
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

	Log.Info("load config: ")
	cfgBytes, err := utils.PrintAsJSON(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(*cfgBytes))
	Log.Debug("debug message is enabled")

	//fmt.Println(lastDate, firstDate, bin)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.POSTGRES_TIMEOUT)
	defer cancel()

	dsn := "postgres://" + cfg.POSTGRES_USER + ":" + cfg.POSTGRES_PASSWORD + "@" + cfg.POSTGRES_HOST + ":" + cfg.POSTGRES_PORT + "/" + cfg.POSTGRES_DB + "?sslmode=disable"
	storage, err := storage.NewStorage(ctx, dsn, Log)
	if err != nil {
		Log.Error("not init postgres storage")
		panic(err)
	}

	pgxTransaction, err := storage.Db.Begin(storage.Ctx)
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

	// определяем новые транзакции для каждого пользователя
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
	Log.Info("messages", slog.Int("count", len(messages)))

	// отправляем операции в брокер
	if len(messages) == 0 {
		Log.Info("no new updates for users")
		err = pgxTransaction.Commit(ctx)
		if err != nil {
			Log.Error("Error commit all db changes:", slog.String("err", err.Error()))
		} else {
			Log.Info("DB changes committed")
		}
		storage.Close()
		return
	}
	err = kofd_updater_services.SendToNats(cfg, Log, messages)
	if err != nil {
		Log.Error("Error broker send:", slog.String("err", err.Error()))
		//storage.Close()
		//return
	}

	err = pgxTransaction.Commit(ctx)
	if err != nil {
		Log.Error("Error commit all db changes:", slog.String("err", err.Error()))
	} else {
		Log.Info("DB changes committed")
	}

	storage.Close()

}
