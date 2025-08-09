package summary

import (
	"context"
	"log/slog"
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
	log *slog.Logger) (modelsI.TypeTransactionsTotal, error) {

	op := "summary.getSummaryDate"
	log = log.With(slog.String("op", op))

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
	log *slog.Logger, cfg *config.Config) (string, *models.InlineKeyboardMarkup, error) {

	op := "summary.getAllChecks"
	log = log.With(slog.String("op", op))

	markups := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}
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
	for _, cheque := range data {

		//+ " " +
		//		cheque.ChequeJSON[0].Name + " /р. " + cheque.ChequeJSON[0].Size.String +
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
	}

	log.Info("transactions count", slog.Int("count", len(data)))

	if len([]rune(sb.String())) > 4096 {
		return "сообщение слишком большое, сократите период", markups, err
	}

	return sb.String(), markups, nil
}
