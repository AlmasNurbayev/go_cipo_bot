package summary

import (
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	tele "gopkg.in/telebot.v4"
)

var (
	SummaryMenu = &tele.ReplyMarkup{ResizeKeyboard: true}

	BtnDay   = SummaryMenu.Text("день")
	BtnWeek  = SummaryMenu.Text("неделя")
	BtnMonth = SummaryMenu.Text("месяц")
)

func Init(b *tele.Bot, storage *storage.Storage,
	log *slog.Logger, cfg *config.Config) {
	SummaryMenu.Reply(
		SummaryMenu.Row(BtnDay, BtnWeek, BtnMonth),
	)

	// /summary
	b.Handle("/summary", func(c tele.Context) error {
		return c.Send("Выберите период:", SummaryMenu)
	})

	// Обработка кнопок
	b.Handle(tele.OnText, func(c tele.Context) error {
		switch c.Text() {
		case BtnDay.Text:
			return CurentDay(b, c, storage, log, cfg)
		case BtnWeek.Text:
			return c.Send("Сводка за неделю")
		case BtnMonth.Text:
			return c.Send("Сводка за месяц")
		default:
			return nil
		}
	})
}
