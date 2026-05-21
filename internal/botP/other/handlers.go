package other

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func otherSiteParserJSONlogHandler(log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.otherSiteParserJSONlogHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		cb := update.Message
		err := utils.SendAction(ctx, cb.Chat.ID, "typing", b)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
		}
		logs, err := ParseAppLog(cfg.SITE_PARSER_JSON_LOG_PATH, 800)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      "Ошибка чтения логов: " + err.Error(),
				ParseMode: models.ParseModeHTML,
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return

		}
		var txt strings.Builder
		txt.WriteString("<b>Запуски JSON парсера сайта:</b> \n")
		for _, item := range logs {
			statusColor := ""
			switch item.Status {
			case "success":
				statusColor = "🟢 "
			case "error":
				statusColor = "🔴 "
			default:
				statusColor = "🟡 "
			}
			imgStr := ""
			if item.IsContainImages {
				imgStr = "🖼️ "
			}
			txt.WriteString(statusColor + imgStr + item.Date + " / " + item.BasePrefix + " / " + fmt.Sprint(item.CountQnt) + "\n")
		}
		log.Info("other called button", slog.String("text", msg.Text))
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      txt.String(),
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}

	}
}
