package config

import (
	"testing"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func TestGetSettingsInt(t *testing.T) {
	settings := []models.SettingsEntity{
		{Key: "SUMMARY_CHECKS_MAX_DAYS", Value: []any{float64(7)}},
		{Key: "EMPTY", Value: []any{}},
		{Key: "STRING", Value: []any{"неделя"}},
		{Key: "FLOAT", Value: []any{7.9}},
	}

	tests := []struct {
		name string
		key  string
		def  int
		want int
	}{
		{"значение из настройки", "SUMMARY_CHECKS_MAX_DAYS", 7, 7},
		{"ключа нет - берем умолчание", "MISSING", 7, 7},
		{"ключа нет - умолчание другое", "MISSING", 3, 3},
		{"пустое значение", "EMPTY", 7, 7},
		{"значение не число", "STRING", 7, 7},
		{"дробное отбрасывается", "FLOAT", 1, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSettingsInt(tt.key, settings, tt.def); got != tt.want {
				t.Errorf("GetSettingsInt(%q, def=%d) = %d, ожидалось %d",
					tt.key, tt.def, got, tt.want)
			}
		})
	}
}

func TestGetSettingsIntNoSettings(t *testing.T) {
	if got := GetSettingsInt("SUMMARY_CHECKS_MAX_DAYS", nil, 7); got != 7 {
		t.Errorf("на пустых настройках вернулось %d, ожидалось 7", got)
	}
}
