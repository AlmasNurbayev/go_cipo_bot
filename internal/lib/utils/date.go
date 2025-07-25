package utils

import (
	"strings"
	"time"
)

func GetDateByMode(mode string) (time.Time, time.Time) {
	parts := strings.Split(mode, " ")

	start, end := time.Now(), time.Now()

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

	}
	return start, end
}
