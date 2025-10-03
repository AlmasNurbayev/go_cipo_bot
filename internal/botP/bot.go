package botP

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP/charts"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP/finance"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP/middleware"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP/qnt"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP/summary"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotApp struct {
	Log      *slog.Logger
	Bot      *bot.Bot
	Cfg      *config.Config
	Storage  *storage.Storage
	Ctx      context.Context
	Settings []modelsI.SettingsEntity
}

func NewApp(ctx context.Context, cfg *config.Config, log *slog.Logger) (*BotApp, error) {
	dsn := "postgres://" + cfg.POSTGRES_USER + ":" + cfg.POSTGRES_PASSWORD + "@" + cfg.POSTGRES_HOST + ":" + cfg.POSTGRES_INT_PORT + "/" + cfg.POSTGRES_DB + "?sslmode=disable"

	ctxDB, cancel := context.WithTimeout(context.Background(), cfg.POSTGRES_TIMEOUT)
	defer cancel()

	storage, err := storage.NewStorage(ctxDB, dsn, log)
	if err != nil {
		log.Error("not init postgres storage")
		panic(err)
	}

	settings, err := config.GetSettings(storage, *log)
	if err != nil {
		log.Error("not init settings from db")
		panic(err)
	}

	opts := []bot.Option{
		bot.WithMiddlewares(middleware.CheckUser(storage, log)),
		bot.WithDefaultHandler(defaultHandler),
		bot.WithCheckInitTimeout(cfg.BOT_TIMEOUT),
	}

	b, err := bot.New(cfg.BOT_TOKEN, opts...)
	if err != nil {
		log.Error("error create bot instance", slog.String("err", err.Error()))
		return nil, err
	}

	return &BotApp{
		Log:      log,
		Bot:      b,
		Storage:  storage,
		Cfg:      cfg,
		Ctx:      ctx,
		Settings: settings,
	}, nil
}

func (b *BotApp) Run() {
	summary.Init(b.Bot, b.Storage, b.Log, b.Cfg)
	charts.Init(b.Bot, b.Storage, b.Log, b.Cfg)
	qnt.Init(b.Bot, b.Storage, b.Log, b.Cfg)
	finance.Init(b.Bot, b.Storage, b.Log, b.Cfg, b.Settings)
	b.Log.Info("bot try started", slog.String("port", "8443"))
	b.Bot.Start(b.Ctx)
}

func (b *BotApp) Stop() {
	_, err := b.Bot.Close(b.Ctx)
	if err != nil {
		b.Log.Error("error closing bot", slog.String("err", err.Error()))
	}
	b.Storage.Close()
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Сообщение не распознано",
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		fmt.Println("error sending message", err.Error())
	}
}
