package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

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

	// kafka, err := botP.NewKafkaReader(ctx, cfg, Log, botApp.Bot, botApp.Storage)
	// if err != nil {
	// 	Log.Error("error create kafka app", slog.String("err", err.Error()))
	// 	panic(err)
	// }

	go func() {
		botApp.Run()
	}()
	go func() {
		httpApp.Run()
	}()

	go func() {
		if err := botP.RunNatsConsumer(ctx, cfg, Log, botApp.Bot, botApp.Storage); err != nil {
			Log.Error("error run nats consumer", slog.String("err", err.Error()))
			cancel()
		}
	}()

	<-ctx.Done()
	Log.Warn("received signal DONE signal")
	fmt.Println("received signal DONE signal")
	botApp.Stop()
	httpApp.Stop()

	time.Sleep(cfg.BOT_TIMEOUT / 2) // timeout)
	Log.Warn("bot, http server, nats stopped")
	Log.Info("============ end bot ============")
}
