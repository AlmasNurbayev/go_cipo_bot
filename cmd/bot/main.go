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
	"github.com/kr/pretty"
)

func main() {
	pretty.Log("reading config...")
	var configEnv string
	flag.StringVar(&configEnv, "configEnv", "", "Path to env-file")
	flag.Parse()

	cfg := config.Mustload(configEnv)
	Log := logger.InitLogger(cfg.ENV, cfg.LOG_ERROR_PATH)
	Log.Info("=== start bot ===")

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
	Log.Info("=== end bot ===")
}

// нужно оборачивать все горутины, которые могут паниковать и не имеют перехвата и лога паники
// для безопасного запуска горутины с логированием паник и ошибок, и вызовом cancel()
// func safeGoCancel(ctx context.Context, cancel context.CancelFunc, log *slog.Logger, name string, fn func() error) {
// 	go func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				log.Error("panic recovered",
// 					slog.String("goroutine", name),
// 					slog.Any("panic", r),
// 					slog.String("stack", string(debug.Stack())),
// 				)
// 				cancel() // аварийное завершение приложения
// 			}
// 		}()
// 		if err := fn(); err != nil {
// 			log.Error("goroutine error",
// 				slog.String("goroutine", name),
// 				slog.String("error", err.Error()),
// 			)
// 			cancel()
// 		}
// 	}()
// }
