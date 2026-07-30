package other

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/botP/middleware"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
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

// otherSendTestMessageHandler показывает список пользователей из БД,
// чтобы выбрать, кому отправить тестовое сообщение
func otherSendTestMessageHandler(storage storageI, log1 *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "other.otherSendTestMessageHandler"
		msg := update.Message
		if msg == nil {
			return
		}
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", msg.From.ID)),
			slog.String("user name", msg.From.Username))

		// команда доступна только администраторам
		isAdmin, err := checkAdmin(ctx, log, storage, msg.From.ID)
		if err != nil {
			sendText(ctx, b, log, msg.Chat.ID, "🔴 Ошибка проверки пользователя: "+err.Error())
			return
		}
		if !isAdmin {
			log.Warn("denied: not admin")
			sendText(ctx, b, log, msg.Chat.ID, "🟡 Команда доступна только администраторам")
			return
		}

		err = utils.SendAction(ctx, msg.Chat.ID, "typing", b)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
		}

		users, err := storage.ListUsers(ctx)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			sendText(ctx, b, log, msg.Chat.ID, "🔴 Ошибка чтения пользователей: "+err.Error())
			return
		}
		if len(users) == 0 {
			sendText(ctx, b, log, msg.Chat.ID, "🟡 Пользователи не найдены")
			return
		}

		log.Info("other called button", slog.String("text", msg.Text), slog.Int("users", len(users)))
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      msg.Chat.ID,
			Text:        "Кому отправить тестовое сообщение?",
			ReplyMarkup: buildUsersInlineKb(users),
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}
	}
}

// otherSendTestCallbackHandler отправляет тестовое сообщение выбранному
// пользователю и возвращает отправителю статус отправки
func otherSendTestCallbackHandler(storage storageI, log1 *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "other.otherSendTestCallbackHandler"
		if update.CallbackQuery == nil {
			return
		}
		cb := update.CallbackQuery
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", cb.From.ID)),
			slog.String("user name", cb.From.Username))

		// сообщение с кнопкой могло быть удалено или устареть - тогда чат недоступен
		chatID, ok := utils.UpdateChatID(update)
		if !ok {
			log.Warn("callback без доступного чата, пропускаем", slog.String("data", cb.Data))
			return
		}
		log.Info("called callback", slog.String("data", cb.Data))
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cb.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}

		// callback-запросы не проверяются в middleware.CheckUser, проверяем сами
		isAdmin, err := checkAdmin(ctx, log, storage, cb.From.ID)
		if err != nil {
			sendText(ctx, b, log, chatID, "🔴 Ошибка проверки пользователя: "+err.Error())
			return
		}
		if !isAdmin {
			log.Warn("denied: not admin")
			sendText(ctx, b, log, chatID, "🟡 Команда доступна только администраторам")
			return
		}

		userId, err := parseSendTestCallback(cb.Data)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			sendText(ctx, b, log, chatID, "🔴 Ошибка: "+err.Error())
			return
		}

		users, err := storage.ListUsers(ctx)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			sendText(ctx, b, log, chatID, "🔴 Ошибка чтения пользователей: "+err.Error())
			return
		}
		var target *modelsI.UserEntity
		for _, user := range users {
			if user.Id == userId {
				target = &user
				break
			}
		}
		if target == nil {
			log.Warn("user not found", slog.Int64("user_id", userId))
			sendText(ctx, b, log, chatID, "🟡 Пользователь не найден (id "+strconv.FormatInt(userId, 10)+")")
			return
		}

		sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    target.Telegram_id,
			Text:      buildTestMessageText(cb.From.Username, time.Now()),
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			log.Error("error sending test message", slog.String("err", err.Error()),
				slog.String("telegram_id", target.Telegram_id), slog.Int64("user_id", target.Id))
			sendText(ctx, b, log, chatID, "🔴 Не отправлено "+target.Telegram_name+
				" (tg "+target.Telegram_id+")\nОшибка: "+err.Error())
			return
		}

		log.Info("test message sent", slog.String("telegram_id", target.Telegram_id),
			slog.Int("message_id", sent.ID))
		sendText(ctx, b, log, chatID, "🟢 Отправлено "+target.Telegram_name+
			" (tg "+target.Telegram_id+"), message_id "+strconv.Itoa(sent.ID))
	}
}

// checkAdmin проверяет, что пользователь есть в БД и имеет роль admin
func checkAdmin(ctx context.Context, log *slog.Logger, storage storageI, telegramId int64) (bool, error) {
	user, err := middleware.GetUserByTelegramId(ctx, log, storage, strconv.FormatInt(telegramId, 10))
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.Role == "admin", nil
}

// sendText отправляет пользователю текст и логирует ошибку отправки
func sendText(ctx context.Context, b *bot.Bot, log *slog.Logger, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		log.Error("error sending message", slog.String("err", err.Error()))
	}
}
