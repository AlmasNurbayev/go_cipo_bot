package main

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/kofd_updater/kofd_updater_services"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/logger"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
)

func main() {
	fmt.Println("Hello, KOFD Updater!")
	fmt.Println("reading config...")
	var (
		configEnv string
		firstDate string
		lastDate  string
		bin       string
	)
	flag.StringVar(&configEnv, "configEnv", "", "Path to env-file")
	flag.StringVar(&firstDate, "firstDate", "", "Date in format YYYY-MM-DD")
	flag.StringVar(&lastDate, "lastDate", "", "Date in format YYYY-MM-DD")
	flag.StringVar(&bin, "bin", "", "BIN of organization")
	flag.Parse()

	cfg := config.Mustload(configEnv)
	Log := logger.InitLogger(cfg.ENV)
	Log.Info("============ start bot ============")

	Log.Info("load config: ")
	cfgBytes, err := utils.PrintAsJSON(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(*cfgBytes))
	Log.Debug("debug message is enabled")

	fmt.Println(lastDate, firstDate, bin)

	dsn := "postgres://" + cfg.POSTGRES_USER + ":" + cfg.POSTGRES_PASSWORD + "@" + cfg.POSTGRES_HOST + ":" + cfg.POSTGRES_PORT + "/" + cfg.POSTGRES_DB + "?sslmode=disable"
	storage, err := storage.NewStorage(dsn, Log, cfg.POSTGRES_TIMEOUT)
	if err != nil {
		Log.Error("not init postgres storage")
		panic(err)
	}

	token, err := kofd_updater_services.GetToken(storage, Log, bin, cfg)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		return
	}

	_, err = kofd_updater_services.GetOperationsFromApi(storage, cfg, Log, bin, token, firstDate, lastDate)
	if err != nil {
		Log.Error("error: ", slog.String("err", err.Error()))
		return
	}

}
