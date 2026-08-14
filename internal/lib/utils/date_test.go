package utils

import "testing"

func TestIsPeriodWithinDays(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		days     int
		want     bool
	}{
		{"один день в лимите недели", "2026-08-14_2026-08-14", 7, true},
		{"ровно неделя", "2026-08-08_2026-08-14", 7, true},
		{"восемь дней", "2026-08-07_2026-08-14", 7, false},
		{"месяц", "2026-08-01_2026-08-31", 7, false},
		{"год", "2026-01-01_2026-12-31", 7, false},
		{"два дня при лимите в день", "2026-08-13_2026-08-14", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := GetPeriodByString(tt.interval)
			if err != nil {
				t.Fatalf("GetPeriodByString(%q): %v", tt.interval, err)
			}
			if got := IsPeriodWithinDays(start, end, tt.days); got != tt.want {
				t.Errorf("IsPeriodWithinDays(%q, %d) = %v, ожидалось %v",
					tt.interval, tt.days, got, tt.want)
			}
		})
	}
}
