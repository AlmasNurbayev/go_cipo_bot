package finance

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	botP "github.com/AlmasNurbayev/go_cipo_bot/internal/botP/api"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
	"github.com/go-analyze/charts"
	"github.com/kr/pretty"
)

func financeOPIUService(ctx context.Context, log1 *slog.Logger, storage *storage.Storage,
	mode string, googleApiKey string,
) ([]byte, string, error) {

	op := "finance.financeOPIUService"
	log := log1.With(slog.String("op", op), slog.String("mode", mode))
	var result []byte
	var text string

	// Получаем границы текущего дня в локальном времени
	start, end, err := utils.GetPeriodByMode(mode)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, text, err
	}

	// Получаем настройки из базы
	settings, err := config.GetSettings(ctx, storage, *log)
	if err != nil {
		log.Error("error getting settings: ", slog.String("err", err.Error()))
		return result, text, err
	}

	books := config.GetSettingsGSheetsSources("FINANCE_GHEETS_SOURCES", settings)
	finance_opiu_special_items := config.GetSettingsString("FINANCE_OPIU_SPECIAL_ITEMS", settings)
	finance_opiu_cost_items := config.GetSettingsString("FINANCE_OPIU_COST_ITEMS", settings)
	finance_opiu_revenue_items := config.GetSettingsString("FINANCE_OPIU_REVENUE_ITEMS", settings)
	finance_usd_rate, err := config.GetSettingsUSDRates("FINANCE_USD_RATE", settings)
	if err != nil {
		log.Error("error getting finance_usd_rate: ", slog.String("err", err.Error()))
		return result, text, err
	}
	finance_planning_margin, err := config.GetSettingsFloat64("FINANCE_PLANNING_MARGIN", settings)
	if err != nil {
		log.Error("error getting FINANCE_PLANNING_MARGIN: ", slog.String("err", err.Error()))
		return result, text, err
	}

	// Получаем данные из гугл таблиц
	sumData := []models.GSheetsEntityV1{}
	for _, v := range books {
		log.Info("book", slog.String("book", v.Book))
		data, err := botP.GsheetsData(googleApiKey, v.Book, v.Sheet, v.Range, log1)
		sumData = append(sumData, data...) // объединяем данные из всех таблиц
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return result, text, err
		}
	}

	// получаем список периодов для отчета, помесячно, между start и end
	periods := []time.Time{}
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	for !current.After(end) {
		if current.Equal(start) || current.After(start) {
			periods = append(periods, current)
		}
		current = current.AddDate(0, 1, 0) // следующий месяц
	}

	// фильтруем данные по периоду и группируем в категории
	groupOpiuData := groupOpiuData(
		sumData,
		periods,
		finance_opiu_revenue_items,
		finance_opiu_cost_items,
		finance_opiu_special_items,
		finance_usd_rate,
	)

	// формируем массивы с заголовками и параметрами
	header := []string{
		"Показатель",
	}
	textAlign := []string{
		charts.AlignLeft,
	}
	var spans []int
	if len(groupOpiuData) < 3 {
		spans = append(spans, 7)
	} else if len(groupOpiuData) < 6 {
		spans = append(spans, 4)
	} else {
		spans = append(spans, 3)
	}
	var totalProfit, totalProfitMargin, totalCosts, totalRevenue,
		totalPrivateCosts, totalZakup int

	// проходим по всем периодам из данных Gsheets
	// если пустая выручка в Gsheets, то пытаемся получить транзакции из касс
	for indexV, v := range groupOpiuData {
		// если выручки нет, то пытаемся получить транзакции из касс
		if v.Revenue.Sum == 0 {
			// получаем начальные и конечные даты месяца
			nextMonth := v.Period.AddDate(0, 1, 0)
			end := nextMonth.Add(-time.Nanosecond)
			// получаем транзакции из касс
			kassas, err := storage.ListKassa(ctx)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				return result, text, err
			}
			data, err := storage.ListTransactionsByDate(ctx, v.Period, end)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				return result, text, err
			}
			summary := utils.ConvertTransToTotal(data, kassas)
			groupOpiuData[indexV].Revenue.Categories = append(v.Revenue.Categories, CategorySum{
				Category: "Сумма чеков по кассам",
				Sum:      summary.SumSales - summary.SumReturns,
			})
			groupOpiuData[indexV].Revenue.Sum += summary.SumSales - summary.SumReturns

			// если есть плановая рентабельность, то считаем себестоимость товара
			if finance_planning_margin > 0 {
				costGoods := (groupOpiuData[indexV].Revenue.Sum / finance_planning_margin)
				groupOpiuData[indexV].Costs.Categories = append(v.Costs.Categories, CategorySum{
					Category: "Прогнозная себестоимость",
					Sum:      costGoods,
				})
				groupOpiuData[indexV].Costs.Sum += costGoods
				groupOpiuData[indexV].Profit = int(groupOpiuData[indexV].Revenue.Sum - groupOpiuData[indexV].Costs.Sum)
				if groupOpiuData[indexV].Revenue.Sum != 0 {
					groupOpiuData[indexV].ProfitMargin = int((float64(groupOpiuData[indexV].Profit) / groupOpiuData[indexV].Revenue.Sum) * 100)
				}
			}
		}

	}

	// заполняем заголовки периодов
	for _, v := range groupOpiuData {
		header = append(header, v.Period.Format("2006-01"))
		textAlign = append(textAlign, charts.AlignRight)
		if len(groupOpiuData) < 3 {
			spans = append(spans, 3)
		} else if len(groupOpiuData) < 6 {
			spans = append(spans, 2)
		} else {
			spans = append(spans, 1)
		}

		// заодно делаем считаем суммы
		totalProfit += v.Profit
		totalCosts += int(v.Costs.Sum)
		totalRevenue += int(v.Revenue.Sum)
		indexPrivate := slices.IndexFunc(v.Special.Categories, func(s CategorySum) bool {
			return s.Category == "Личные нужды"
		})
		if indexPrivate != -1 {
			totalPrivateCosts += int(v.Special.Categories[indexPrivate].Sum)
		}
		indexZakup := slices.IndexFunc(v.Special.Categories, func(s CategorySum) bool {
			return s.Category == "Закуп товаров"
		})
		if indexZakup != -1 {
			totalZakup += int(v.Special.Categories[indexZakup].Sum)
		}
	}
	if totalRevenue != 0 {
		totalProfitMargin = int(float64(totalProfit) / float64(totalRevenue) * 100)
	}

	data := [][]string{}
	// заголовки таблицы данных
	rowRevenue := []string{"Выручка"}
	rowCosts := []string{"Затраты"}
	rowCostsCategory := [][]string{}
	rowProfit := []string{"Прибыль"}
	rowProfitMargin := []string{"Рентабельность"}
	rowSpecial := []string{"Вне отчета"}
	rowSpecialCategory := [][]string{}

	// создаем вложенные подкатегории заранее по всем периодам
	for _, dataPeriod := range groupOpiuData {
		for _, v := range dataPeriod.Costs.Categories {
			if !slices.ContainsFunc(rowCostsCategory, func(s []string) bool { return s[0] == v.Category }) {
				rowCostsCategory = append(rowCostsCategory, []string{v.Category})
			}
		}
		for _, v := range dataPeriod.Special.Categories {
			if !slices.ContainsFunc(rowSpecialCategory, func(s []string) bool { return s[0] == v.Category }) {
				rowSpecialCategory = append(rowSpecialCategory, []string{v.Category})
			}
		}
	}

	// формируем строки таблицы данных
	for _, dataPeriod := range groupOpiuData {

		rowRevenue = append(rowRevenue, utils.FormatNumber(dataPeriod.Revenue.Sum))
		rowCosts = append(rowCosts, utils.FormatNumber(dataPeriod.Costs.Sum))

		// сортируем массив расходов по убыванию
		slices.SortFunc(dataPeriod.Costs.Categories,
			func(a, b CategorySum) int {
				if a.Sum > b.Sum {
					return -1
				}
				if a.Sum < b.Sum {
					return 1
				}
				return 0
			})
		// создаем подкатегории и добавляем в них данные
		for i, v := range rowCostsCategory {
			indexCategory := slices.IndexFunc(dataPeriod.Costs.Categories, func(s CategorySum) bool { return v[0] == s.Category })
			if indexCategory != -1 {
				rowCostsCategory[i] = append(rowCostsCategory[i], utils.FormatNumber(dataPeriod.Costs.Categories[indexCategory].Sum))
			} else {
				rowCostsCategory[i] = append(rowCostsCategory[i], utils.FormatNumber(0))
			}
		}
		// создаем подкатегории и добавляем в них данные
		for i, v := range rowSpecialCategory {
			indexCategory := slices.IndexFunc(dataPeriod.Special.Categories, func(s CategorySum) bool { return v[0] == s.Category })
			if indexCategory != -1 {
				rowSpecialCategory[i] = append(rowSpecialCategory[i], utils.FormatNumber(dataPeriod.Special.Categories[indexCategory].Sum))
			} else {
				rowSpecialCategory[i] = append(rowSpecialCategory[i], utils.FormatNumber(0))
			}
		}
		rowProfit = append(rowProfit, utils.FormatNumber(float64(dataPeriod.Profit)))
		rowProfitMargin = append(rowProfitMargin, utils.FormatNumber(float64(dataPeriod.ProfitMargin))+"%")
		rowSpecial = append(rowSpecial, utils.FormatNumber(dataPeriod.Special.Sum))
	}

	// добавляем маркеры для строк второго уровня
	for i := 0; i < len(rowCostsCategory); i++ {
		rowCostsCategory[i][0] = " • " + rowCostsCategory[i][0]
	}
	for i := 0; i < len(rowSpecialCategory); i++ {
		rowSpecialCategory[i][0] = " • " + rowSpecialCategory[i][0]
	}

	// соединяем показатели
	data = append(data, rowRevenue)
	data = append(data, rowCosts)
	data = append(data, rowCostsCategory...)
	data = append(data, rowProfit)
	data = append(data, rowProfitMargin)
	data = append(data, rowSpecial)
	data = append(data, rowSpecialCategory...)

	//pretty.Log(data)
	//pretty.Log(groupOpiuData)
	if len(data) == 0 {
		return result, text, errors.New("нет данных")
	}

	// генерируем таблицу
	var widthCol int
	if len(data[0]) < 4 {
		widthCol = 200
	} else {
		widthCol = 150
	}

	p := charts.NewPainter(charts.PainterOptions{
		OutputFormat: charts.ChartOutputPNG,
		Width:        len(data[0]) * widthCol, // ширина на каждый период
		Height:       len(data) * 55,          // высота на каждый показатель
	})
	p.FilledRect(0, 0, 810, 300, charts.ColorWhite, charts.ColorWhite, 0.0)
	tableOpt1 := charts.TableChartOption{
		Header:                header,
		HeaderBackgroundColor: charts.ColorRGB(80, 80, 80),
		HeaderFontColor:       charts.ColorRGB(255, 255, 255),
		TextAligns:            textAlign,
		Data:                  data,
		Spans:                 spans,
		Padding: charts.Box{
			Top:    10,
			Right:  20,
			Bottom: 10,
			Left:   10,
		},
		CellModifier: func(tc charts.TableCell) charts.TableCell {
			// форматируем отдельные строки
			if tc.Text == "Прибыль" || tc.Text == "Рентабельность" {
				tc.FontStyle.FontColor = charts.ColorRGB(26, 8, 66)
			}
			if strings.Contains(tc.Text, " • ") {
				tc.FontStyle.FontSize = 11
			}
			return tc
		},
	}

	err = p.TableChart(tableOpt1)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, text, err
	}
	// выгружаем в байты картинки
	result, err = p.Bytes()
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, text, err
	}

	// формируем текст
	text = "<b>Суммарно</b> Выручка " + utils.FormatNumber(float64(totalRevenue)) + "\n"
	text += "Прибыль: " + utils.FormatNumber(float64(totalProfit)) + ", Рентабельность: " +
		utils.FormatNumber(float64(totalProfitMargin)) + "% \n"
	text += "Личные расходы " + utils.FormatNumber(float64(totalPrivateCosts)) +
		", дельта от прибыли " + utils.FormatNumber(float64(totalProfit-totalPrivateCosts)) + "\n"
	text += "Закупки всего " + utils.FormatNumber(float64(totalZakup))

	return result, text, err
}

func financeChartService(ctx context.Context, log1 *slog.Logger, storage *storage.Storage,
	mode string, googleApiKey string,
) ([]byte, string, error) {

	op := "finance.financeChartService"
	log := log1.With(slog.String("op", op), slog.String("mode", mode))
	var result []byte
	var text string
	var err error

	// Получаем границы текущего дня в локальном времени
	// start, end, err := utils.GetPeriodByMode(mode)
	// if err != nil {
	//log.Error("error: ", slog.String("err", err.Error()))
	//return result, text, err
	//}

	// Получаем настройки из базы
	settings, err := config.GetSettings(ctx, storage, *log)
	if err != nil {
		log.Error("error getting settings: ", slog.String("err", err.Error()))
		return result, text, err
	}

	books := config.GetSettingsGSheetsSources("FINANCE_GHEETS_SOURCES", settings)
	finance_opiu_special_items := config.GetSettingsString("FINANCE_OPIU_SPECIAL_ITEMS", settings)
	finance_opiu_cost_items := config.GetSettingsString("FINANCE_OPIU_COST_ITEMS", settings)
	finance_opiu_revenue_items := config.GetSettingsString("FINANCE_OPIU_REVENUE_ITEMS", settings)
	finance_usd_rate, err := config.GetSettingsUSDRates("FINANCE_USD_RATE", settings)
	if err != nil {
		log.Error("error getting finance_usd_rate: ", slog.String("err", err.Error()))
		return result, text, err
	}
	finance_planning_margin, err := config.GetSettingsFloat64("FINANCE_PLANNING_MARGIN", settings)
	if err != nil {
		log.Error("error getting FINANCE_PLANNING_MARGIN: ", slog.String("err", err.Error()))
		return result, text, err
	}

	// Получаем все данные из гугл таблиц
	sumData := []models.GSheetsEntityV1{}
	for _, v := range books {
		log.Info("book", slog.String("book", v.Book))
		data, err := botP.GsheetsData(googleApiKey, v.Book, v.Sheet, v.Range, log1)
		sumData = append(sumData, data...) // объединяем данные из всех таблиц
		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return result, text, err
		}
	}

	// получаем начало каждого месяца текущего года и добавляем в массив
	periods := []time.Time{}
	var labels []string
	now := time.Now()
	year := now.Year()
	location := now.Location()
	for month := time.January; month <= time.December; month++ {
		date := time.Date(year, month, 1, 0, 0, 0, 0, location)
		periods = append(periods, date)
		labels = append(labels, date.Format("2006-01"))
	}

	// фильтруем данные по периоду и группируем в категории
	groupOpiuData := groupOpiuData(
		sumData,
		periods,
		finance_opiu_revenue_items,
		finance_opiu_cost_items,
		finance_opiu_special_items,
		finance_usd_rate,
	)
	//var totalProfit, totalProfitMargin, totalCosts, totalRevenue,
	//	totalPrivateCosts, totalZakup int

	// проходим по всем периодам из данных Gsheets
	// если пустая выручка в Gsheets, то пытаемся получить транзакции из касс
	for indexV, v := range groupOpiuData {
		// если выручки нет, то пытаемся получить транзакции из касс
		if v.Revenue.Sum == 0 {
			// получаем начальные и конечные даты месяца
			nextMonth := v.Period.AddDate(0, 1, 0)
			end := nextMonth.Add(-time.Nanosecond)
			// получаем транзакции из касс
			kassas, err := storage.ListKassa(ctx)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				return result, text, err
			}
			data, err := storage.ListTransactionsByDate(ctx, v.Period, end)
			if err != nil {
				log.Error("error: ", slog.String("err", err.Error()))
				return result, text, err
			}
			summary := utils.ConvertTransToTotal(data, kassas)
			groupOpiuData[indexV].Revenue.Categories = append(v.Revenue.Categories, CategorySum{
				Category: "Сумма чеков по кассам",
				Sum:      summary.SumSales - summary.SumReturns,
			})
			groupOpiuData[indexV].Revenue.Sum += summary.SumSales - summary.SumReturns

			// если есть плановая рентабельность, то считаем себестоимость товара
			if finance_planning_margin > 0 {
				costGoods := (groupOpiuData[indexV].Revenue.Sum / finance_planning_margin)
				groupOpiuData[indexV].Costs.Categories = append(v.Costs.Categories, CategorySum{
					Category: "Прогнозная себестоимость",
					Sum:      costGoods,
				})
				groupOpiuData[indexV].Costs.Sum += costGoods
				groupOpiuData[indexV].Profit = int(groupOpiuData[indexV].Revenue.Sum - groupOpiuData[indexV].Costs.Sum)
				if groupOpiuData[indexV].Revenue.Sum != 0 {
					groupOpiuData[indexV].ProfitMargin = int((float64(groupOpiuData[indexV].Profit) / groupOpiuData[indexV].Revenue.Sum) * 100)
				}
			}
		}
	}
	values := make([][]float64, 3)
	for i := range groupOpiuData {
		values[0] = append(values[0], groupOpiuData[i].Revenue.Sum)
		values[1] = append(values[1], groupOpiuData[i].Costs.Sum)
		values[2] = append(values[2], float64(groupOpiuData[i].Profit))
	}
	pretty.Log(values)

	opt := charts.NewBarChartOptionWithData(values)
	opt.Title.Text = "Крайние месяцы, этот год и предыдущий"
	opt.XAxis.Labels = labels
	opt.XAxis.LabelRotation = charts.DegreesToRadians(45)
	opt.XAxis.LabelFontStyle.FontSize = 8
	opt.YAxis[0].SplitLineShow = charts.Ptr(true)
	//opt.YAxis[1].Min = charts.Ptr(0.0)

	opt.Legend = charts.LegendOption{
		SeriesNames: []string{
			"выручка", "затраты", "прибыль",
		},
		Offset: charts.OffsetRight,
	}
	opt.SeriesLabelPosition = charts.PositionTop

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
		return result, text, err
	}

	result, err = p.Bytes()
	if err != nil {
		log.Error("error creating bar chart", slog.String("err", err.Error()))
		return result, text, err
	}

	return result, text, err
}

type opiuColumn struct {
	Period  time.Time
	Revenue struct {
		Sum        float64
		Categories []CategorySum
	}
	Costs struct {
		Sum        float64
		Categories []CategorySum
	}
	Profit       int
	ProfitMargin int
	Special      struct {
		Sum        float64
		Categories []CategorySum
	}
}

type CategorySum struct {
	Category string
	Sum      float64
	SumUSD   float64
}

// группируем данные по периодам и категориям
func groupOpiuData(
	data []models.GSheetsEntityV1,
	periods []time.Time,
	revenueItems []string,
	costItems []string,
	specialItems []string,
	usdRates []models.USDRates,
) []opiuColumn {

	// для быстрого поиска
	inSet := func(cat string, set []string) bool {
		return slices.Contains(set, cat)
	}

	result := []opiuColumn{}

	for _, period := range periods {
		col := opiuColumn{Period: period}

		// временные карты для агрегации по категориям
		revenueMap := make(map[string]CategorySum)
		costMap := make(map[string]CategorySum)
		specialMap := make(map[string]CategorySum)

		for _, d := range data {
			if !sameMonth(d.Period, period) {
				continue
			}

			switch {
			case inSet(d.Category, revenueItems):
				col.Revenue.Sum += d.Sum
				cs := revenueMap[d.Category]
				cs.Category = d.Category
				cs.Sum += d.Sum
				cs.SumUSD += d.SumUSD
				revenueMap[d.Category] = cs

			case inSet(d.Category, costItems):
				col.Costs.Sum += d.Sum
				cs := costMap[d.Category]
				cs.Category = d.Category
				cs.Sum += d.Sum
				cs.SumUSD += d.SumUSD
				costMap[d.Category] = cs

			case inSet(d.Category, specialItems):
				cs := specialMap[d.Category]
				cs.Category = d.Category
				if d.SumUSD != 0 {
					// если есть закуп в USD, то находим курс нужного года и добавляем
					indexUSDYear := slices.IndexFunc(usdRates, func(s models.USDRates) bool {
						return s.Year == period.Year()
					})
					if indexUSDYear != -1 {
						d.Sum += float64(d.SumUSD * float64(usdRates[indexUSDYear].Rate))
					}
				}
				cs.Sum += d.Sum
				cs.SumUSD += d.SumUSD
				specialMap[d.Category] = cs
				col.Special.Sum += d.Sum
			}
		}

		// сконвертировать map -> slice
		for _, v := range revenueMap {
			col.Revenue.Categories = append(col.Revenue.Categories, v)
		}
		for _, v := range costMap {
			col.Costs.Categories = append(col.Costs.Categories, v)
		}
		for _, v := range specialMap {
			col.Special.Categories = append(col.Special.Categories, v)
		}

		col.Profit = int(col.Revenue.Sum - col.Costs.Sum)

		if col.Revenue.Sum != 0 {
			col.ProfitMargin = int((float64(col.Profit) / col.Revenue.Sum) * 100)
		}

		// если прибыль или внешние расходы нулевые, то данных нет и не добавляем период
		if col.Profit == 0 && col.Special.Sum == 0 {
			continue
		}

		result = append(result, col)
	}
	return result
}

// вспомогательная функция: сравнение по месяцу
func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}
