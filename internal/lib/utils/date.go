package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

// получаем границы периода из текста формата "2025-01-01_2025-01-31"
func GetPeriodByString(interval string) (time.Time, time.Time, error) {
	parts := strings.Split(interval, "_")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, errors.New("неверный формат дат")
	}
	var start time.Time
	var end time.Time
	start, err := time.Parse("2006-01-02", parts[0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err = time.Parse("2006-01-02", parts[1])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	// начало суток
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	// конец суток
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, int(time.Nanosecond*999999999), end.Location())
	return start, end, nil
}

// получаем границы периода из текста формата "итог тек. день"
func GetPeriodByMode(mode string) (time.Time, time.Time, error) {
	parts := strings.Split(mode, " ")
	if len(parts) < 2 {
		return time.Time{}, time.Time{}, errors.New("неверный формат дат")
	}

	var start time.Time
	var end time.Time

	if parts[1] == "тек." && parts[2] == "день" {
		now := time.Now()
		loc := now.Location()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end = start.Add(24 * time.Hour).Add(-time.Nanosecond)
		//return start, end
	} else if parts[1] == "пр." && parts[2] == "день" {
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour).Add(-time.Nanosecond)
		//return start, end
	} else if parts[1] == "тек." && parts[2] == "неделя" {
		now := time.Now()
		location := now.Location()
		weekday := int(now.Weekday())
		if weekday == 0 { // воскресенье = 7
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -weekday+1)
		end = start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	} else if parts[1] == "пр." && parts[2] == "неделя" {
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 { // воскресенье → 7
			weekday = 7
		}
		startOfThisWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -weekday+1)
		start = startOfThisWeek.AddDate(0, 0, -7)
		end = start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	} else if parts[1] == "тек." && parts[2] == "месяц" {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	} else if parts[1] == "пр." && parts[2] == "месяц" {
		now := time.Now()
		startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start = startOfThisMonth.AddDate(0, -1, 0)
		end = startOfThisMonth.Add(-time.Nanosecond)
	} else if parts[1] == "тек." && parts[2] == "квартал" {
		now := time.Now()
		month := now.Month()
		quarter := (int(month)-1)/3 + 1
		startMonth := time.Month((quarter-1)*3 + 1)
		start = time.Date(now.Year(), startMonth, 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 3, 0).Add(-time.Nanosecond)
	} else if parts[1] == "пр." && parts[2] == "квартал" {
		now := time.Now()
		month := int(now.Month())
		quarter := (month-1)/3 + 1
		startMonthOfCurrentQuarter := time.Month((quarter-1)*3 + 1)
		startOfCurrentQuarter := time.Date(now.Year(), startMonthOfCurrentQuarter, 1, 0, 0, 0, 0, now.Location())
		start = startOfCurrentQuarter.AddDate(0, -3, 0)
		end = startOfCurrentQuarter.Add(-time.Nanosecond)
	} else if parts[1] == "тек." && parts[2] == "год" {
		now := time.Now()
		start = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	} else if parts[1] == "пр." && parts[2] == "год" {
		now := time.Now()
		location := now.Location()
		start = time.Date(now.Year()-1, time.January, 1, 0, 0, 0, 0, location)
		end = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, location).Add(-time.Nanosecond)
	} else if []rune(parts[1])[0] == '2' && len(parts) == 2 {
		// если второе слово начинается с 2 и передано 2 слова, то распарсиваем как год
		input := strings.Join(parts[1:], " ")
		t, err := time.Parse("2006", input)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("неверный формат дат YYYY")
		}
		start = time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	} else if []rune(parts[1])[0] == '2' && len(parts) == 3 {
		// если второе слово начинается с 2 и передано 3 слова, то распарсиваем как год и месяц
		input := strings.Join(parts[1:], " ")
		t, err := time.Parse("2006 01", input)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("неверный формат дат YYYY MM")
		}
		start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	} else if []rune(parts[1])[0] == '2' && len(parts) == 4 {
		// если второе слово начинается с 2 и передано 4 слова, то распарсиваем как год, месяц и день
		input := strings.Join(parts[1:], " ")
		t, err := time.Parse("2006 01 02", input)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("неверный формат дат YYYY MM DD")
		}
		start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 1).Add(-time.Nanosecond)
	} else {
		return time.Time{}, time.Time{}, errors.New("неизвестный формат даты")
	}
	return start, end, nil
}

// GetLastDaysPeriod возвращает начало и конец периода, массив с датами последних N дней
// где N - количество дней, переданное в параметре days
func GetLastDaysPeriod(days int) (time.Time, time.Time, []models.TransactionsByDays, error) {

	now := time.Now()
	// начало суток (N-1) дней назад
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		AddDate(0, 0, -days+1)
	// конец сегодняшнего дня
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
	var dataDays []models.TransactionsByDays
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -i) // от текущей даты назад
		dataDays = append(dataDays, models.TransactionsByDays{
			Date:  time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location()),
			Day:   d.Day(),
			Month: int(d.Month()),
			Year:  d.Year(),
			Count: 0,
			Sum:   0,
		})
	}

	return start, end, dataDays, nil
}

// GetLastDaysPeriodPrevYear возвращает начало и конец периода год назад, массив с датами этих дней
// где N - количество дней, переданное в параметре days
func GetLastDaysPeriodPrevYear(days int) (time.Time, time.Time, []models.TransactionsByDays, error) {

	now := time.Now()
	// начало суток (N-1) дней назад
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		AddDate(0, 0, -days+1)
	// конец сегодняшнего дня
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())

	// сдвигаем на год назад
	start = start.AddDate(-1, 0, 0)
	end = end.AddDate(-1, 0, 0)

	var dataDays []models.TransactionsByDays
	for i := 0; i < days; i++ {
		d := now.AddDate(-1, 0, -i) // от текущей даты назад
		dataDays = append(dataDays, models.TransactionsByDays{
			Date:  time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location()),
			Day:   d.Day(),
			Month: int(d.Month()),
			Year:  d.Year(),
			Count: 0,
			Sum:   0,
		})
	}

	return start, end, dataDays, nil
}

// GetCurrentYearPeriod возвращает начало и конец периода, массив с месяцами текущего года
// где N - количество дней, переданное в параметре monthes
func GetCurrentYearPeriod(monthes int) (time.Time, time.Time, []models.TransactionsByDays, error) {

	now := time.Now()
	// начало этого года
	start := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())

	// конец этого года
	end := time.Date(now.Year()+1, time.January, 1, 0, 0, 0, 0, now.Location()).
		Add(-time.Nanosecond)

	var dataMonthes []models.TransactionsByDays
	year := now.Year()
	for m := now.Month(); m >= time.January; m-- {
		d := time.Date(year, m, 1, 0, 0, 0, 0, now.Location())
		dataMonthes = append(dataMonthes, models.TransactionsByDays{
			Date:  time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location()),
			Day:   d.Day(),
			Month: int(d.Month()),
			Year:  d.Year(),
			Count: 0,
			Sum:   0,
		})
	}
	return start, end, dataMonthes, nil
}

// GetPrevYearPeriod возвращает начало и конец периода, массив с месяцами прошлого года
// где N - количество дней, переданное в параметре monthes
func GetPrevYearPeriod(monthes int) (time.Time, time.Time, []models.TransactionsByDays, error) {

	now := time.Now()
	// Начало прошлого года
	start := time.Date(now.Year()-1, time.January, 1, 0, 0, 0, 0, now.Location())

	// Конец прошлого года (31 декабря 23:59:59.999999999)
	end := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()).
		Add(-time.Nanosecond)

	var dataMonthes []models.TransactionsByDays
	year := now.AddDate(-1, 0, 0).Year()
	for m := now.Month(); m >= time.January; m-- {
		d := time.Date(year, m, 1, 0, 0, 0, 0, now.Location())
		dataMonthes = append(dataMonthes, models.TransactionsByDays{
			Date:  time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location()),
			Day:   d.Day(),
			Month: int(d.Month()),
			Year:  d.Year(),
			Count: 0,
			Sum:   0,
		})
	}
	return start, end, dataMonthes, nil
}

func GetLastMonthesPeriod(monthes int) (time.Time, time.Time, []models.TransactionsByMonthes, error) {

	now := time.Now()
	start := time.Date(
		now.Year(),
		now.Month(),
		1,
		0, 0, 0, 0,
		now.Location(),
	).AddDate(0, -monthes, 0)
	// конец сегодняшнего дня
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
	var dataMonthes []models.TransactionsByMonthes

	for i := range monthes {
		d := start.AddDate(0, i, 0) // от текущей даты вперед
		dataMonthes = append(dataMonthes, models.TransactionsByMonthes{
			Month: int(d.Month()),
			Year:  d.Year(),
			Count: 0,
			Sum:   0,
		})
	}
	// добавляем последний месяц
	dataMonthes = append(dataMonthes, models.TransactionsByMonthes{
		Month: int(end.Month()),
		Year:  end.Year(),
		Count: 0,
		Sum:   0,
	})
	//slices.Reverse(dataMonthes)

	return start, end, dataMonthes, nil
}
