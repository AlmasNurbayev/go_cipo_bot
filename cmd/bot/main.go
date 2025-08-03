package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/logger"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
)

func main() {
	fmt.Println("reading config...")
	var configEnv string
	flag.StringVar(&configEnv, "configEnv", "", "Path to env-file")
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	botApp, err := botP.NewApp(ctx, cfg, Log)
	if err != nil {
		Log.Error("error create bot app", slog.String("err", err.Error()))
		panic(err)
	}
	httpApp, err := botP.NewHttpApp(cfg, Log)
	if err != nil {
		Log.Error("error create http app", slog.String("err", err.Error()))
		panic(err)
	}

	kafka, err := botP.NewKafkaReader(ctx, cfg, Log, botApp.Bot, botApp.Storage)
	if err != nil {
		Log.Error("error create kafka app", slog.String("err", err.Error()))
		panic(err)
	}

	go func() {
		botApp.Run()
	}()
	go func() {
		httpApp.Run()
	}()
	go func() {
		kafka.Run()
	}()

	//done := make(chan os.Signal, 1)
	//signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-ctx.Done()
	Log.Info("received signal DONE signal")
	fmt.Println("received signal DONE signal")

	botApp.Stop()
	httpApp.Stop()
	Log.Warn("bot, http server, kafka stopped")
}
