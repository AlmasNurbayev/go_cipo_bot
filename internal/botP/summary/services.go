package summary

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot/models"
)

type storageI interface {
	ListTransactionsByDate(context.Context, time.Time, time.Time) ([]modelsI.TransactionEntity, error)
	GetTransactionById(context.Context, int64) (modelsI.TransactionEntity, error)
	ListKassa(context.Context) ([]modelsI.KassaEntity, error)
}

func getSummaryDate(ctx context.Context, mode string, storage storageI,
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

	kassas, err := storage.ListKassa(ctx)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	data, err := storage.ListTransactionsByDate(ctx, start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	log.Info("transactions count", slog.Int("count", len(data)))

	result = utils.ConvertTransToTotal(data, kassas)
	result.StartDate = start
	result.EndDate = end
	result.DateMode = mode

	// вытаскиваем активные кассы чтобы получить остатки денег
	for i := range result.KassaTotal {
		if !result.KassaTotal[i].IsActive {
			continue
		}
		// операции в обратном порядке
		for j := len(data) - 1; i >= 0; i-- {
			if data[j].Kassa_id == result.KassaTotal[i].KassaId {
				result.KassaTotal[i].CashAmount = data[j].Availablesum.Float64
				break
			}
		}
	}

	return result, err
}

func getAllChecksService(ctx context.Context, mode string, storage storageI,
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

	data, err := storage.ListTransactionsByDate(ctx, start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", markups, err
	}

	var sb strings.Builder

	var keyboardButtons []models.InlineKeyboardButton

	for index, cheque := range data {

		typeOperation := utils.GetTypeOperationText(cheque)
		if typeOperation == "Возврат" {
			typeOperation = "⚠️Возврат"
		}
		typePayment := utils.GetTypePaymentText(cheque)
		if typePayment == "Неизвестно" {
			typePayment = ""
		}

		sb.WriteString("<b>" + strconv.Itoa(index+1) + ". ")
		sb.WriteString(cheque.Kassa_name.String + " - " + strings.TrimSpace(cheque.Operationdate.Time.Format("15:04")) +
			" - " + typeOperation + " " + typePayment + " " + utils.FormatNumber(cheque.Sum_operation.Float64) + " ₸")
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
		if len(keyboardButtons) == 8 {
			inlineKeyboard = append(inlineKeyboard, keyboardButtons)
			keyboardButtons = []models.InlineKeyboardButton{}
		}
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

func getAnalyticsService(ctx context.Context, mode string, storage storageI,
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
	diff := int(end.Sub(start).Hours() / 24)

	data, err := storage.ListTransactionsByDate(ctx, start, end)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", markups, err
	}
	log.Info("transactions count", slog.Int("count", len(data)))

	var seasons []modelsI.Simple
	var days []modelsI.Simple
	var monthes []modelsI.Simple
	var vids []modelsI.Simple
	var kassas []modelsI.Simple

	// собираем данные в массивы
	totalOperationsSum := 0.0
	totalSum := 0.0
	totalDiscount := 0.0
	totalReturns := 0.0
	for _, item := range data {
		if item.Type_operation != 1 {
			continue
		}

		// получаем сумму чеков
		if item.Subtype.Int64 == 3 {
			totalOperationsSum -= item.Sum_operation.Float64
		} else {
			totalOperationsSum += item.Sum_operation.Float64
		}

		// разбиваем по именам из JSON
		for _, cheque := range item.ChequeJSON {
			if item.Subtype.Int64 == 3 {
				// Возврат, минусом
				totalSum -= cheque.Sum
				totalReturns += cheque.Sum
				seasons = append(seasons, modelsI.Simple{Item: cheque.Season.String, Count: 1, Sum: -cheque.Sum})
				days = append(days, modelsI.Simple{Item: item.Operationdate.Time.Format("02.01.2006"), Count: 1, Sum: -cheque.Sum})
				monthes = append(monthes, modelsI.Simple{Item: item.Operationdate.Time.Format("2006.01"), Count: 1, Sum: -cheque.Sum})
				vids = append(vids, modelsI.Simple{Item: cheque.VidModeli.String, Count: 1, Sum: -cheque.Sum})
				kassas = append(kassas, modelsI.Simple{Item: item.Kassa_name.String, Count: 1, Sum: -cheque.Sum})
				totalDiscount -= (cheque.NominalPrice - cheque.DiscountPrice) * float64(cheque.Qnt)
			} else {
				totalSum += cheque.Sum
				seasons = append(seasons, modelsI.Simple{Item: cheque.Season.String, Count: 1, Sum: cheque.Sum})
				days = append(days, modelsI.Simple{Item: item.Operationdate.Time.Format("02.01.2006"), Count: 1, Sum: cheque.Sum})
				monthes = append(monthes, modelsI.Simple{Item: item.Operationdate.Time.Format("2006.01"), Count: 1, Sum: cheque.Sum})
				vids = append(vids, modelsI.Simple{Item: cheque.VidModeli.String, Count: 1, Sum: cheque.Sum})
				kassas = append(kassas, modelsI.Simple{Item: item.Kassa_name.String, Count: 1, Sum: cheque.Sum})
				totalDiscount += (cheque.NominalPrice - cheque.DiscountPrice) * float64(cheque.Qnt)
			}
		}
	}

	// группируем по Item и сортируем
	days = utils.GroupByItem(days)
	sort.Slice(days, func(i, j int) bool {
		return days[i].Item < days[j].Item
	})
	monthes = utils.GroupByItem(monthes)
	sort.Slice(monthes, func(i, j int) bool {
		return monthes[i].Item < monthes[j].Item
	})
	seasons = utils.GroupByItem(seasons)
	sort.Slice(seasons, func(i, j int) bool {
		return seasons[i].Sum > seasons[j].Sum // убывание
	})
	vids = utils.GroupByItem(vids)
	sort.Slice(vids, func(i, j int) bool {
		return vids[i].Sum > vids[j].Sum // убывание
	})
	kassas = utils.GroupByItem(kassas)
	sort.Slice(kassas, func(i, j int) bool {
		return kassas[i].Sum > kassas[j].Sum // убывание
	})

	var sb strings.Builder
	sb.WriteString("<b>" + mode + "</b>\n")
	sb.WriteString("В аналитику не попали чеки на: " + utils.FormatNumber(totalOperationsSum-totalSum) + " ₸\n\n")
	sb.WriteString("<b>Сезоны:</b>\n")
	sb.WriteString(utils.StructToString(seasons, false, true) + "\n\n")
	if diff <= 60 {
		sb.WriteString("<b>Дни:</b>\n")
		sb.WriteString(utils.StructToString(days, true, true) + "\n\n")
	}
	sb.WriteString("<b>Месяцы:</b>\n")
	sb.WriteString(utils.StructToString(monthes, false, true) + "\n\n")
	sb.WriteString("<b>Виды:</b>\n")
	sb.WriteString(utils.StructToString(vids, false, true) + "\n\n")
	sb.WriteString("<b>Кассы:</b>\n")
	sb.WriteString(utils.StructToString(kassas, false, true) + "\n\n")

	sb.WriteString("<b>Примененные скидки </b>от " + utils.FormatNumber(totalOperationsSum) + "\n")
	sb.WriteString(" = " + utils.FormatNumber(totalDiscount) + " ₸ или " +
		utils.FormatNumber(totalDiscount/totalSum*100) + "% \n")

	sb.WriteString("<b>Возвраты </b> ")
	sb.WriteString(" = " + utils.FormatNumber(totalReturns) + " ₸ или " +
		utils.FormatNumber(totalReturns/totalSum*100) + "% \n")

	if len([]rune(sb.String())) > 4096 {
		return "сообщение слишком большое, сократите период", markups, err
	}

	return sb.String(), markups, nil
}

func getOneCheckService(ctx context.Context, queryString string, storage storageI,
	log1 *slog.Logger, cfg *config.Config) ([]models.InputMedia, string, error) {

	op := "summary.getOneCheck"
	log := log1.With(slog.String("op", op))

	var modelsToSend []models.InputMediaPhoto
	//var inlineKeyboard [][]models.InlineKeyboardButton
	var checkID int64

	queryArr := strings.Split(queryString, "_")
	if len(queryArr) < 2 {
		return nil, "", errors.New("неверный формат запроса чека, должен быть в формате: getCheck_1234567890")
	}
	checkID, err := strconv.ParseInt(queryArr[1], 10, 64)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return nil, "", err
	}

	//log.Info("queryString", slog.String("queryString", queryString))

	data, err := storage.GetTransactionById(ctx, checkID)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return nil, "", err
	}
	typeOperation := utils.GetTypeOperationText(data)
	if typeOperation == "Возврат" {
		typeOperation = "⚠️Возврат"
	}
	typePayment := utils.GetTypePaymentText(data)
	if typePayment == "Неизвестно" {
		typePayment = ""
	}

	var sb strings.Builder
	sb.WriteString("<b>чек №" + strconv.FormatInt(data.Id, 10) + " от " + data.Operationdate.Time.Format("2006.01.02 15:04") + "</b>\n")
	sb.WriteString("касса: " + data.Kassa_name.String + "\n")
	sb.WriteString("тип операции: " + typeOperation + " " + typePayment + ", сумма: " + utils.FormatNumber(data.Sum_operation.Float64) + "\n")

	if len(data.ChequeJSON) > 0 {
		sb.WriteString("товары: ")
		for _, item := range data.ChequeJSON {
			sb.WriteString("\n • " + item.Name + " (" + item.Size.String + ") ₸ " + utils.FormatNumber(item.Sum))
			if item.VidModeli.String != "" && item.Season.String != "" {
				sb.WriteString(" (" + item.VidModeli.String + ", " + item.Season.String + ")")
			}
			if item.NominalPrice-item.DiscountPrice != 0 {
				sb.WriteString(" - скидка " + utils.FormatNumber(item.NominalPrice-item.DiscountPrice) + " ₸")
			}
			if item.Qnt > 1 {
				sb.WriteString(" x" + strconv.Itoa(int(item.Qnt)))
			}
			if item.MainImageURL.String != "" {
				modelsToSend = append(modelsToSend, models.InputMediaPhoto{
					Media: item.MainImageURL.String,
				})
			}
		}
	}

	if len(modelsToSend) > 0 {
		modelsToSend[0].Caption = sb.String()
		modelsToSend[0].ParseMode = models.ParseModeHTML
	}

	if len([]rune(sb.String())) > 1096 {
		return nil, "", errors.New("сообщение слишком большое, сократите период")
	}

	var inputMedia []models.InputMedia
	for _, media := range modelsToSend {
		inputMedia = append(inputMedia, &media)
	}

	// Если есть фото, то отправляем их, иначе просто текст
	if len(inputMedia) == 0 {
		//fmt.Println(sb.String())
		return nil, sb.String(), nil
	} else {
		return inputMedia, sb.String(), nil
	}
}

func getFullTextCheckService(ctx context.Context, queryString string, storage storageI,
	log1 *slog.Logger) (string, error) {

	op := "summary.getAllChecks"
	log := log1.With(slog.String("op", op))

	var checkID int64

	queryArr := strings.Split(queryString, "_")
	if len(queryArr) < 2 {
		return "", errors.New("неверный формат запроса чека, должен быть в формате: getFullTextCheck_1234567890")
	}
	checkID, err := strconv.ParseInt(queryArr[1], 10, 64)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}

	log.Info("queryString", slog.String("queryString", queryString))

	data, err := storage.GetTransactionById(ctx, checkID)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return "", err
	}
	result := data.Cheque.String

	return result, nil
}
