package summary

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func summaryHandler(storage storageI,
	log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryButtonHandler"
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.Message.From.ID)), slog.String("user name", update.Message.From.Username))
		msg := update.Message
		if msg == nil {
			return
		}
		//return
		log.Info("summary called button", slog.String("text", msg.Text))
		parts := strings.Split(msg.Text, " ")
		if len(parts) < 2 {
			log.Warn("summary called < 2 words: " + msg.Text)
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "запрос итоги должен быть в формате: 'итоги тек. день' или 'итоги пр. день' и т.д. или 'итоги 2024 08' или 'итоги 2024 08 21' или 'итоги 2024'",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}

		if msg.Text == "итоги произ. дата" {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text: "отправьте сообщение в формате 'итоги год месяц', например 'итоги 2024 08'\n" +
					"или 'итоги 2024 08 21' для получения итогов по конкретному дню",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
			return
		}

		data, err := getSummaryDate(ctx, msg.Text, storage, log)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка получения данных",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
		p := message.NewPrinter(language.Russian)
		text := "<b>" + data.DateMode +
			" (" + data.StartDate.Format("02.01.2006") + " - " + data.EndDate.Format("02.01.2006") + ")</b> \n" +
			"количество чеков: " + strconv.Itoa(data.Count) + "\n" +
			"чистая сумма продаж: <b>" + p.Sprintf("%.0f", data.Sum) + "</b> \n" +
			" в т.ч. кеш: " + p.Sprintf("%.0f", data.SumSalesCash-data.SumReturnsCash) + "\n" +
			"        карта: " + p.Sprintf("%.0f", data.SumSalesCard-data.SumReturnsCard) + "\n" +
			"        смешанно: " + p.Sprintf("%.0f", data.SumSalesMixed-data.SumReturnsMixed) + "\n" +
			"        прочее: " + p.Sprintf("%.0f", data.SumSalesOther-data.SumReturnsOther) + "\n" +
			"Cумма продаж: " + p.Sprintf("%.0f", data.SumSales) + "\n" +
			"Cумма возвратов: " + p.Sprintf("%.0f", data.SumReturns) + "\n" +
			"\n"

		// если есть больше 1 кассы, то выводим информацию по ним
		if len(data.KassaTotal) > 1 {
			text += "по кассам:\n"
		}

		for _, kassa := range data.KassaTotal {
			// если нет чеков по кассе или одна касса, то пропускаем
			if kassa.Count == 0 || len(data.KassaTotal) == 1 {
				continue
			}
			text += "<b>" + kassa.NameKassa + "</b> (" + kassa.NameOrganization + ") " +
				"кол-во чеков: " + strconv.Itoa(kassa.Count) + "\n" +
				"чистая сумма продаж: " + p.Sprintf("%.0f", kassa.Sum) + "\n" +
				"сумма продаж: " + p.Sprintf("%.0f", kassa.SumSales) + "\n" +
				"сумма возвратов: " + p.Sprintf("%.0f", kassa.SumReturns) + "\n"
		}
		text +=
			"Выемки: " + p.Sprintf("%.0f", data.SumOutputCash) + ", " +
				"Внесения: " + p.Sprintf("%.0f", data.SumInputCash) + "\n" +
				"Наличие денег в кассах:"

		for _, kassa := range data.KassaTotal {
			if kassa.CashAmount != 0 {
				text += kassa.NameKassa +
					" " + p.Sprintf("%.0f", kassa.CashAmount) + "\n"
			}
		}

		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: summaryInlineKb(data.StartDate, data.EndDate, checksMaxDays(ctx, storage, log)),
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}
	}
}

func summaryInlineKb(data1 time.Time, data2 time.Time, maxDays int) *models.InlineKeyboardMarkup {
	//if strings.Contains(text, "день") || strings.Contains(text, "неделя") {
	start := data1.Format("2006-01-02")
	end := data2.Format("2006-01-02")

	keyboard := [][]models.InlineKeyboardButton{}
	// списки чеков отдаем только за короткий период - за месяц это сотни строк
	if utils.IsPeriodWithinDays(data1, data2, maxDays) {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: "🔍 Все чеки", CallbackData: "summary_allChecks_" + start + "_" + end},
			{Text: "Все чеки #", CallbackData: "summary_allChecksTable_" + start + "_" + end},
		})
	}
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: "Аналитика", CallbackData: "summary_analytics_" + start + "_" + end},
		{Text: "Диаграмма по дням", CallbackData: "summary_chartsDay_" + start + "_" + end},
	})

	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func summaryCallbackHandler(storage storageI,
	log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryCallbackHandler"
		if update.CallbackQuery == nil {
			return
		}
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username))
		cb := update.CallbackQuery
		// сообщение с кнопкой могло быть удалено или устареть - тогда чат недоступен
		chatID, ok := utils.UpdateChatID(update)
		if !ok {
			log.Warn("callback без доступного чата, пропускаем", slog.String("data", cb.Data))
			return
		}
		log.Info("called callback", slog.String("data", cb.Data))
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}

		if strings.Contains(cb.Data, "summary_allChecksTable_") {
			rows, err := getAllChecksTableService(ctx, cb.Data, storage, log)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   checksErrorText(err),
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			if len(rows) == 0 {
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "за этот период чеков нет",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			parts := splitTableRows(rows, richMessageMaxLen, maxRowsPerRichTable)
			log.Info("sending checks table", slog.Int("rows", len(rows)), slog.Int("parts", len(parts)))
			for i, part := range parts {
				if i > 0 {
					time.Sleep(400 * time.Millisecond)
				}
				_, err = b.SendRichMessage(ctx, &bot.SendRichMessageParams{
					ChatID:      chatID,
					RichMessage: models.InputRichMessage{Markdown: renderChecksTable(part)},
					ReplyMarkup: tableKeyboard(part),
				})
				if err == nil {
					continue
				}
				// rich-сообщения появились в Bot API 10.1 - если сервер их не знает,
				// показываем ту же таблицу моноширинным блоком
				log.Error("error sending rich message, отправляю запасным вариантом",
					slog.String("err", err.Error()))
				sendChecksTablePre(ctx, b, chatID, part, log)
			}
			return
		}

		if strings.Contains(cb.Data, "summary_allChecks_") {
			chunks, err := getAllChecksService(ctx, cb.Data, storage, log, cfg)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   checksErrorText(err),
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			if len(chunks) == 0 {
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "за этот период чеков нет",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			log.Info("sending checks", slog.Int("parts", len(chunks)))
			for i, chunk := range chunks {
				if i > 0 {
					// лимит Telegram - примерно 1 сообщение в секунду в чат
					time.Sleep(400 * time.Millisecond)
				}
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:      chatID,
					Text:        chunk.Text,
					ParseMode:   models.ParseModeHTML,
					ReplyMarkup: chunk.Markup,
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()),
						slog.Int("part", i+1))
				}
			}
		}

		if strings.Contains(cb.Data, "summary_analytics_") {
			err := utils.SendAction(ctx, chatID, "typing", b)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
			}
			response, markups, err := getAnalyticsService(ctx, cb.Data, storage, log, cfg)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Ошибка получения данных",
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				return
			}
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        response,
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: markups,
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
		//if cb.Data == "summary_Day" {

	}
}

// checksErrorText - понятную пользователю причину показываем текстом,
// остальное общей формулировкой
func checksErrorText(err error) string {
	var tooLong errPeriodTooLong
	if errors.As(err, &tooLong) {
		return tooLong.Error()
	}
	return "Ошибка получения данных"
}

const (
	// лимит текста rich-сообщения - 32768 символов, берем с запасом
	richMessageMaxLen = 30000
	// строк на одну таблицу: и блоков в rich-сообщении не больше 500, и
	// кнопки чеков не должны занимать пол-экрана - 40 строк это 5 рядов по 8
	maxRowsPerRichTable = 40
)

var checksTableHeader = checkRow{
	Num:      "№",
	DateTime: "дата - время",
	Type:     "тип",
	Sum:      "сумма",
	Card:     "в т.ч. карта",
	Cash:     "в т.ч. нал",
}

func tableCells(row checkRow) [6]string {
	return [6]string{row.Num, row.DateTime, row.Type, row.Sum, row.Card, row.Cash}
}

// renderChecksTable собирает таблицу в GFM-разметке. Колонки дополняются
// пробелами по самому широкому значению: rich-сообщение отрисует настоящую
// таблицу, а в запасном моноширинном варианте колонки не разъедутся.
func renderChecksTable(rows []checkRow) string {
	// номер и суммы прижимаем вправо, дату и тип - влево
	alignRight := [6]bool{true, false, false, true, true, true}

	var widths [6]int
	for i, cell := range tableCells(checksTableHeader) {
		widths[i] = utf8.RuneCountInString(cell)
	}
	for _, row := range rows {
		for i, cell := range tableCells(row) {
			if width := utf8.RuneCountInString(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}

	var sb strings.Builder
	writeRow := func(row checkRow) {
		for i, cell := range tableCells(row) {
			// GFM пробелы по краям ячейки игнорирует, а моноширинному
			// запасному варианту они и дают выравнивание
			pad := strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell))
			if alignRight[i] {
				sb.WriteString("| " + pad + cell + " ")
			} else {
				sb.WriteString("| " + cell + pad + " ")
			}
		}
		sb.WriteString("|\n")
	}

	writeRow(checksTableHeader)
	for i, right := range alignRight {
		if right {
			sb.WriteString("|" + strings.Repeat("-", widths[i]+1) + ":")
		} else {
			sb.WriteString("|:" + strings.Repeat("-", widths[i]+1))
		}
	}
	sb.WriteString("|\n")
	for _, row := range rows {
		writeRow(row)
	}

	return sb.String()
}

// tableKeyboard - кнопки-номера на чеки этой части таблицы, каждая открывает
// чек целиком. Номера совпадают с колонкой "№" своего сообщения.
func tableKeyboard(rows []checkRow) models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton
	var buttons []models.InlineKeyboardButton

	for _, row := range rows {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         row.Num,
			CallbackData: "getCheck_" + strconv.FormatInt(row.ID, 10),
		})
		if len(buttons) == maxButtonsPerRow {
			keyboard = append(keyboard, buttons)
			buttons = nil
		}
	}
	if len(buttons) > 0 {
		keyboard = append(keyboard, buttons)
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

// splitTableRows режет строки на части так, чтобы готовая таблица каждой части
// влезала в limit; maxRows дополнительно ограничивает число строк (0 - без ограничения).
func splitTableRows(rows []checkRow, limit int, maxRows int) [][]checkRow {
	var parts [][]checkRow
	var cur []checkRow

	for _, row := range rows {
		tooManyRows := maxRows > 0 && len(cur) >= maxRows
		if len(cur) > 0 && (tooManyRows || utils.TelegramLen(renderChecksTable(append(cur, row))) > limit) {
			parts = append(parts, cur)
			cur = nil
		}
		cur = append(cur, row)
	}
	if len(cur) > 0 {
		parts = append(parts, cur)
	}

	return parts
}

// sendChecksTablePre - запасной вариант для серверов без Bot API 10.1:
// та же таблица моноширинным блоком в обычном сообщении, где лимит 4096,
// поэтому режем ее еще раз.
func sendChecksTablePre(ctx context.Context, b *bot.Bot, chatID int64,
	rows []checkRow, log *slog.Logger) {

	const preWrapLen = len("<pre></pre>")

	for i, part := range splitTableRows(rows, utils.TelegramMaxMessageLen-preWrapLen, 0) {
		if i > 0 {
			time.Sleep(400 * time.Millisecond)
		}
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "<pre>" + renderChecksTable(part) + "</pre>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: tableKeyboard(part),
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}
	}
}

func summaryGetCheckHandler(storage storageI,
	log1 *slog.Logger, cfg *config.Config) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryGetCheckHandler"
		if update.CallbackQuery == nil {
			return
		}
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username))
		cb := update.CallbackQuery
		// сообщение с кнопкой могло быть удалено или устареть - тогда чат недоступен
		chatID, ok := utils.UpdateChatID(update)
		if !ok {
			log.Warn("callback без доступного чата, пропускаем", slog.String("data", cb.Data))
			return
		}
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}
		err = utils.SendAction(ctx, chatID, "upload_photo", b)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
		}
		inputMedia, stringResponce, err := getOneCheckService(ctx, cb.Data, storage, log, cfg)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Ошибка получения данных",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}

		// Если фото есть, то отправляем МедиаГруппой, иначе просто текстом
		if len(inputMedia) == 0 {
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      stringResponce,
				ParseMode: models.ParseModeHTML,
				ReplyMarkup: &models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{
								Text:         "Полный текст чека",
								CallbackData: "getFullTextCheck_" + strings.Split(cb.Data, "_")[1],
							},
						},
					},
				},
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		} else {
			// если есть фото, отправляем медиа группой
			_, err = b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
				ChatID: chatID,
				Media:  inputMedia,
			})
			// Запоминаем, что кнопку "Полный текст чека" уже отправили
			FullTextButtonIsSending := false
			if err != nil {
				// если не получилось отправить медиа группой, то отправляем текстом
				log.Error("error sending media group", slog.String("err", err.Error()))
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    chatID,
					Text:      stringResponce,
					ParseMode: models.ParseModeHTML,
					ReplyMarkup: &models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{
									Text:         "Полный текст чека",
									CallbackData: "getFullTextCheck_" + strings.Split(cb.Data, "_")[1],
								},
							},
						},
					},
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
				FullTextButtonIsSending = true
			}
			// так как нельзя отправить кнопку к МедиаГруппе, то отправляем отдельным сообщением
			// если ранее такая кнопка уже была отправлена, то не отправляем повторно
			if !FullTextButtonIsSending {
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Полный текст чека",
					ReplyMarkup: &models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{
									Text:         "Открыть",
									CallbackData: "getFullTextCheck_" + strings.Split(cb.Data, "_")[1],
								},
							},
						},
					},
				})
				if err != nil {
					log.Error("error sending message", slog.String("err", err.Error()))
				}
			}
		}
	}
}

func summaryFullTextCheckHandler(storage storageI,
	log1 *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		op := "summary.summaryFullCheckHandler"
		if update.CallbackQuery == nil {
			return
		}
		log := log1.With(slog.String("op", op), slog.Attr(slog.Int64("id", update.CallbackQuery.From.ID)),
			slog.String("user name", update.CallbackQuery.From.Username))
		cb := update.CallbackQuery
		// сообщение с кнопкой могло быть удалено или устареть - тогда чат недоступен
		chatID, ok := utils.UpdateChatID(update)
		if !ok {
			log.Warn("callback без доступного чата, пропускаем", slog.String("data", cb.Data))
			return
		}
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
		if err != nil {
			log.Error("error answering callback query", slog.String("err", err.Error()))
		}

		response, err := getFullTextCheckService(ctx, cb.Data, storage, log)
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Ошибка получения данных",
			})
			if err != nil {
				log.Error("error sending message", slog.String("err", err.Error()))
			}
		}
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      response,
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			log.Error("error sending message", slog.String("err", err.Error()))
		}

	}
}
