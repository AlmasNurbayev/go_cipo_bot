package charts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
)

func charts30Days(ctx context.Context, storage *storage.Storage, log1 *slog.Logger) ([]byte, error) {
	op := "charts.charts30Days"
	log := log1.With(slog.String("op", op))

	start, end, daysArr, err := utils.GetLastDaysPeriod(30)
	if err != nil {
		log.Error("error getting last days period", slog.String("err", err.Error()))
		return nil, err
	}

	data, err := storage.ListTransactionsByDate(ctx, start, end)
	if err != nil {
		log.Error("error listing transactions by date", slog.String("err", err.Error()))
		return nil, err
	}

	var dataDays []models.TransactionsByDays

	for i := range daysArr {
		if len(daysArr[i]) != 3 {
			log.Error("invalid date format in daysArr", slog.Int("index", i), slog.Any("date", daysArr[i]))
			continue
		}
		year := daysArr[i][0]
		month := daysArr[i][1]
		day := daysArr[i][2]

		count := 0
		sum := 0.0

		for _, d := range data {
			if d.Operationdate.Time.Year() == year && int(d.Operationdate.Time.Month()) == month && d.Operationdate.Time.Day() == day {
				count++
				sum += d.Sum_operation.Float64
			}
		}

		dataDays = append(dataDays, models.TransactionsByDays{
			Day:   day,
			Month: month,
			Year:  year,
			Count: count,
			Sum:   sum,
		})
	}
	fmt.Println("dataDays:", dataDays)

	var file []byte

	return file, nil

}
