package summary

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type storageI interface {
	ListTransactionsByDate(context.Context, time.Time, time.Time) ([]modelsI.TransactionEntity, error)
}

func getSummaryDate(mode string, storage storageI,
	log1 *slog.Logger) (modelsI.TypeTransactionsTotal, error) {

	op := "summary.getSummaryDate"
	log := log1.With(slog.String("op", op))

	var result modelsI.TypeTransactionsTotal

	// Получаем границы текущего дня в локальном времени
	start, end, err := utils.GetPeriodByMode(mode)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}

	result.StartDate = start
	result.EndDate = end
	result.DateMode = mode

	data, err := storage.ListTransactionsByDate(context.Background(), start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	log.Info("transactions count", slog.Int("count", len(data)))

	result = utils.ConvertTransToTotal(result, data)

	return result, err
}

func getAllChecks(mode string, b *bot.Bot, storage storageI,
	log1 *slog.Logger, cfg *config.Config) (string, models.InlineKeyboardMarkup, error) {

	op := "summary.getAllChecks"
	log := log1.With(slog.String("op", op))

	var markups models.InlineKeyboardMarkup
	var inlineKeyboard [][]models.InlineKeyboardButton

	mode = strings.ReplaceAll(mode, "summary_allChecks_", "")

	// Получаем границы текущего дня в локальном времени
	start, end, err := utils.GetPeriodByString(mode)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", markups, err
	}
	log.Info("date", slog.Time("start", start), slog.Time("end", end))

	data, err := storage.ListTransactionsByDate(context.Background(), start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", markups, err
	}

	var sb strings.Builder

	var keyboardButtons []models.InlineKeyboardButton

	for index, cheque := range data {

		sb.WriteString("<b>")
		sb.WriteString(cheque.Kassa_name.String + " - " + strings.TrimSpace(cheque.Operationdate.Time.Format("15:04")) +
			" - " + utils.FormatNumber(cheque.Sum_operation.Float64) + " ₸")
		sb.WriteString("\n")
		sb.WriteString("</b>")
		if cheque.ChequeJSON != nil {
			for _, item := range cheque.ChequeJSON {
				sb.WriteString(" • " + item.Name + " (" + item.Size.String + ") ₸ " + utils.FormatNumber(item.Sum) + "\n")
			}
		}
		sb.WriteString("\n")
		keyboardButtons = append(keyboardButtons, models.InlineKeyboardButton{
			Text:         strconv.Itoa(index + 1),
			CallbackData: "getCheck_" + strconv.Itoa(int(cheque.Id)),
		},
		)
	}

	if len([]rune(sb.String())) > 4096 {
		// нужно возвращать пустой markup, иначе не уйдет
		return "сообщение слишком большое, сократите период",
			models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{},
			}, err
	}

	log.Info("transactions count", slog.Int("count", len(data)))

	if len(keyboardButtons) > 0 {
		inlineKeyboard = append(inlineKeyboard, keyboardButtons)
	}
	markups = models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	return sb.String(), markups, nil
}

func getAnalytics(mode string, storage storageI,
	log1 *slog.Logger, cfg *config.Config) (string, *models.InlineKeyboardMarkup, error) {

	op := "summary.getAnalytics"
	log := log1.With(slog.String("op", op))

	markups := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}
	mode = strings.ReplaceAll(mode, "summary_analytics_", "")

	// Получаем границы текущего дня в локальном времени
	start, end, err := utils.GetPeriodByString(mode)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", markups, err
	}
	log.Info("date", slog.Time("start", start), slog.Time("end", end))

	data, err := storage.ListTransactionsByDate(context.Background(), start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", markups, err
	}
	log.Info("transactions count", slog.Int("count", len(data)))

	var seasons []modelsI.Simple
	var days []modelsI.Simple
	var vids []modelsI.Simple

	// собираем данные в массивы
	totalOperationsSum := 0.0
	totalSum := 0.0
	totalDiscount := 0.0
	for _, item := range data {
		if item.Type_operation != 1 {
			continue
		}
		totalOperationsSum += item.Sum_operation.Float64
		for _, cheque := range item.ChequeJSON {
			totalSum += cheque.Sum
			totalDiscount += (cheque.NominalPrice - cheque.DiscountPrice) * float64(cheque.Qnt)
			seasons = append(seasons, modelsI.Simple{Item: cheque.Season.String, Sum: cheque.Sum})
			days = append(days, modelsI.Simple{Item: item.Operationdate.Time.Format("02.01.2006"), Sum: cheque.Sum})
			vids = append(vids, modelsI.Simple{Item: cheque.VidModeli.String, Sum: cheque.Sum})
		}
	}

	// группируем по Item и сортируем
	days = utils.GroupByItem(days)
	sort.Slice(days, func(i, j int) bool {
		return days[i].Item < days[j].Item
	})
	seasons = utils.GroupByItem(seasons)
	sort.Slice(seasons, func(i, j int) bool {
		return seasons[i].Sum > seasons[j].Sum // убывание
	})
	vids = utils.GroupByItem(vids)
	sort.Slice(vids, func(i, j int) bool {
		return vids[i].Sum > vids[j].Sum // убывание
	})

	var sb strings.Builder
	sb.WriteString("<b>" + mode + "</b>\n")
	sb.WriteString("В аналитику не попали чеки на: " + utils.FormatNumber(totalOperationsSum-totalSum) + " ₸\n\n")
	sb.WriteString("<b>Сезоны</b>\n")
	sb.WriteString(utils.StructToString(seasons) + "\n")
	sb.WriteString("<b>Дни</b>\n")
	sb.WriteString(utils.StructToString(days) + "\n")
	sb.WriteString("<b>Виды</b>\n")
	sb.WriteString(utils.StructToString(vids) + "\n\n")
	sb.WriteString("<b>Примененные скидки от " + utils.FormatNumber(totalOperationsSum) + "</b>\n")
	sb.WriteString(" =" + utils.FormatNumber(totalDiscount) + " ₸ или " +
		utils.FormatNumber(totalDiscount/totalSum*100) + "% \n")

	if len([]rune(sb.String())) > 4096 {
		return "сообщение слишком большое, сократите период", markups, err
	}

	return sb.String(), markups, nil
}
