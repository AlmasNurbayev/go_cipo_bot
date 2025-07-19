package bot

import (
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/bot/middleware"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/bot/summary"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	tele "gopkg.in/telebot.v4"
)

type BotApp struct {
	Log     *slog.Logger
	Bot     *tele.Bot
	Cfg     *config.Config
	Storage *storage.Storage
}

func NewApp(cfg *config.Config, log *slog.Logger) (*BotApp, error) {
	dsn := "postgres://" + cfg.POSTGRES_USER + ":" + cfg.POSTGRES_PASSWORD + "@" + cfg.POSTGRES_HOST + ":" + cfg.POSTGRES_PORT + "/" + cfg.POSTGRES_DB + "?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), cfg.POSTGRES_TIMEOUT)
	defer cancel()

	storage, err := storage.NewStorage(ctx, dsn, log)
	if err != nil {
		log.Error("not init postgres storage")
		panic(err)
	}

	pref := tele.Settings{
		Token:  cfg.BOT_TOKEN,
		Poller: &tele.LongPoller{Timeout: cfg.BOT_TIMEOUT},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Error("error create telebot instance", slog.String("err", err.Error()))
		return nil, err
	}
	return &BotApp{
		Log:     log,
		Bot:     b,
		Storage: storage,
		Cfg:     cfg,
	}, nil
}

func (b *BotApp) Run() {
	b.Log.Info("bot started", slog.String("port", "8443"))
	b.Bot.Use(middleware.CheckUser(b.Storage, b.Log, b.Cfg.BOT_TIMEOUT))
	summary.Init(b.Bot, b.Storage, b.Log, b.Cfg)

	b.Bot.Start()
}

func (b *BotApp) Stop() {
	b.Bot.Stop()
	b.Storage.Close()
}
