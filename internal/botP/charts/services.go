package charts

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-analyze/charts"
)

// возвращает данные за последние 30 дней и год назад в виде
// массив байтов PNG, сумму за последние 30 дней, сумму 30 дней за год назад
func charts30Days(ctx context.Context, storage *storage.Storage, log1 *slog.Logger) ([]byte, float64, float64, error) {
	op := "charts.charts30Days"
	log := log1.With(slog.String("op", op))

	days := 30

	// получаем границы и массив-шаблон с датами
	start, end, dataDays, err := utils.GetLastDaysPeriod(days)
	if err != nil {
		log.Error("error getting last days period", slog.String("err", err.Error()))
		return nil, 0, 0, err
	}

	data, err := storage.ListTransactionsByDate(ctx, start, end)
	if err != nil {
		log.Error("error listing transactions by date", slog.String("err", err.Error()))
		return nil, 0, 0, err
	}
	sumCurrent := 0.0
	for index, row := range dataDays {
		for _, d := range data {
			if d.Type_operation != 1 {
				continue
			}
			if d.Operationdate.Time.Year() == row.Year && int(d.Operationdate.Time.Month()) == row.Month && d.Operationdate.Time.Day() == row.Day {
				dataDays[index].Count++
				switch d.Subtype.Int64 {
				case 2:
					dataDays[index].Sum += d.Sum_operation.Float64
					sumCurrent += d.Sum_operation.Float64
				case 3:
					dataDays[index].Sum -= d.Sum_operation.Float64
					sumCurrent -= d.Sum_operation.Float64
				}
			}
		}
	}

	// тоже самое но на предыдущий год
	startPrev, endPrev, dataDaysPrev, err := utils.GetLastDaysPeriodPrevYear(days)
	if err != nil {
		log.Error("error getting last days period", slog.String("err", err.Error()))
		return nil, 0, 0, err
	}
	dataPrev, err := storage.ListTransactionsByDate(ctx, startPrev, endPrev)
	if err != nil {
		log.Error("error listing transactions by date", slog.String("err", err.Error()))
		return nil, 0, 0, err
	}
	sumPrev := 0.0
	for index, row := range dataDaysPrev {
		for _, d := range dataPrev {
			if d.Type_operation != 1 {
				continue
			}
			if d.Operationdate.Time.Year() == row.Year && int(d.Operationdate.Time.Month()) == row.Month && d.Operationdate.Time.Day() == row.Day {
				dataDaysPrev[index].Count++
				switch d.Subtype.Int64 {
				case 2:
					dataDaysPrev[index].Sum += d.Sum_operation.Float64
					sumPrev += d.Sum_operation.Float64
				case 3:
					dataDaysPrev[index].Sum -= d.Sum_operation.Float64
					sumPrev -= d.Sum_operation.Float64
				}
			}
		}
	}

	values := make([][]float64, 2)
	var labels []string
	for i := len(dataDays) - 1; i >= 0; i-- {
		values[0] = append(values[0], dataDays[i].Sum)
		labels = append(labels, strconv.Itoa(dataDays[i].Day))
	}
	for i := len(dataDaysPrev) - 1; i >= 0; i-- {
		values[1] = append(values[1], dataDaysPrev[i].Sum)
	}

	falseShow := false
	opt := charts.NewHorizontalBarChartOptionWithData(values)
	opt.Title.Text = "Крайние 30 дней, сейчас и год назад"
	opt.YAxis.Labels = labels
	opt.Legend = charts.LegendOption{
		SeriesNames: []string{
			dataDays[0].Date.Format("2006.01"), dataDaysPrev[0].Date.Format("2006.01"),
		},
		Offset:       charts.OffsetRight,
		OverlayChart: &falseShow,
	}
	opt.SeriesList[0].MarkLine.AddLines(charts.SeriesMarkTypeAverage)
	opt.SeriesList[1].MarkLine.AddLines(charts.SeriesMarkTypeAverage)
	//opt.SeriesList[0].MarkPoint.AddPoints(charts.SeriesMarkTypeMax, charts.SeriesMarkTypeMin)
	opt.SeriesLabelPosition = charts.PositionTop
	opt.SeriesList[0].Label.ValueFormatter = func(f float64) string {
		return charts.FormatValueHumanizeShort(f, 0, false)
	}
	opt.SeriesList[1].Label.ValueFormatter = func(f float64) string {
		return charts.FormatValueHumanizeShort(f, 0, false)
	}
	show := true
	opt.SeriesList[0].Label.Show = &show
	opt.SeriesList[1].Label.Show = &show
	opt.BarHeight = 10

	p := charts.NewPainter(charts.PainterOptions{
		Width:  600,
		Height: 1000,
	})

	err = p.HorizontalBarChart(opt)
	if err != nil {
		log.Error("error creating bar chart", slog.String("err", err.Error()))
		return nil, 0, 0, err
	}

	buf, err := p.Bytes()
	if err != nil {
		log.Error("error creating bar chart", slog.String("err", err.Error()))
		return nil, 0, 0, err
	}

	return buf, sumCurrent, sumPrev, nil

}

func chartsCurrentYear(ctx context.Context, storage *storage.Storage, log1 *slog.Logger) ([]byte, error) {
	op := "charts.charts12Month"
	log := log1.With(slog.String("op", op))

	monthes := 12

	// получаем границы и массив-шаблон с датами
	start, end, dataMonthes, err := utils.GetCurrentYearPeriod(monthes)
	if err != nil {
		log.Error("error getting last days period", slog.String("err", err.Error()))
		return nil, err
	}

	data, err := storage.ListTransactionsByDate(ctx, start, end)
	if err != nil {
		log.Error("error listing transactions by date", slog.String("err", err.Error()))
		return nil, err
	}
	for index, row := range dataMonthes {
		for _, d := range data {
			if d.Type_operation != 1 {
				continue
			}
			if d.Operationdate.Time.Year() == row.Year && int(d.Operationdate.Time.Month()) == row.Month {
				dataMonthes[index].Count++
				switch d.Subtype.Int64 {
				case 2:
					dataMonthes[index].Sum += d.Sum_operation.Float64
				case 3:
					dataMonthes[index].Sum -= d.Sum_operation.Float64
				}
			}
		}
	}

	// тоже самое но на предыдущий год
	startPrev, endPrev, dataMonthesPrev, err := utils.GetPrevYearPeriod(monthes)
	if err != nil {
		log.Error("error getting last days period", slog.String("err", err.Error()))
		return nil, err
	}
	dataPrev, err := storage.ListTransactionsByDate(ctx, startPrev, endPrev)
	if err != nil {
		log.Error("error listing transactions by date", slog.String("err", err.Error()))
		return nil, err
	}
	for index, row := range dataMonthesPrev {
		for _, d := range dataPrev {
			if d.Type_operation != 1 {
				continue
			}
			if d.Operationdate.Time.Year() == row.Year && int(d.Operationdate.Time.Month()) == row.Month {
				dataMonthesPrev[index].Count++
				switch d.Subtype.Int64 {
				case 2:
					dataMonthesPrev[index].Sum += d.Sum_operation.Float64
				case 3:
					dataMonthesPrev[index].Sum -= d.Sum_operation.Float64
				}
			}
		}
	}

	values := make([][]float64, 2)
	var labels []string
	for i := range dataMonthes {
		values[0] = append(values[0], dataMonthes[i].Sum)
		labels = append(labels, strconv.Itoa(dataMonthes[i].Month))
	}
	for i := range dataMonthesPrev {
		values[1] = append(values[1], dataMonthesPrev[i].Sum)
	}

	opt := charts.NewBarChartOptionWithData(values)
	opt.Title.Text = "Крайние месяцы, этот год и предыдущий"
	opt.XAxis.Labels = labels
	opt.Legend = charts.LegendOption{
		SeriesNames: []string{
			time.Now().Format("2006"),
			time.Now().AddDate(-1, 0, 0).Format("2006"),
		},
		Offset: charts.OffsetRight,
	}
	opt.SeriesList[0].MarkLine.AddLines(charts.SeriesMarkTypeAverage)
	opt.SeriesList[1].MarkLine.AddLines(charts.SeriesMarkTypeAverage)
	opt.SeriesLabelPosition = charts.PositionTop
	// opt.SeriesList[0].Label.ValueFormatter = func(f float64) string {
	// 	return charts.FormatValueHumanizeShort(f, 0, false)
	// }
	// opt.SeriesList[1].Label.ValueFormatter = func(f float64) string {
	// 	return charts.FormatValueHumanizeShort(f, 0, false)
	// }
	show := true
	opt.SeriesList[0].Label.Show = &show
	opt.SeriesList[1].Label.Show = &show
	//opt.BarWidth = 10

	p := charts.NewPainter(charts.PainterOptions{
		Width:  900,
		Height: 600,
	})

	err = p.BarChart(opt)
	if err != nil {
		log.Error("error creating bar chart", slog.String("err", err.Error()))
		return nil, err
	}

	buf, err := p.Bytes()
	if err != nil {
		log.Error("error creating bar chart", slog.String("err", err.Error()))
		return nil, err
	}

	return buf, nil

}
