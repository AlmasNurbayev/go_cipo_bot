package summary

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

// fakeStorage - заглушка storageI, для настроек нужен только GetSettings
type fakeStorage struct {
	settings []modelsI.SettingsEntity
	err      error
}

func (f fakeStorage) GetSettings(context.Context) ([]modelsI.SettingsEntity, error) {
	return f.settings, f.err
}

func (f fakeStorage) ListTransactionsByDate(context.Context, time.Time, time.Time) ([]modelsI.TransactionEntity, error) {
	return nil, nil
}

func (f fakeStorage) GetTransactionById(context.Context, int64) (modelsI.TransactionEntity, error) {
	return modelsI.TransactionEntity{}, nil
}

func (f fakeStorage) ListKassa(context.Context) ([]modelsI.KassaEntity, error) {
	return nil, nil
}

func TestChecksMaxDays(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name    string
		storage fakeStorage
		want    int
	}{
		{
			name: "значение из настройки",
			storage: fakeStorage{settings: []modelsI.SettingsEntity{
				{Key: keyChecksMaxDays, Value: []any{float64(3)}},
			}},
			want: 3,
		},
		{
			name:    "строки настройки нет",
			storage: fakeStorage{settings: []modelsI.SettingsEntity{{Key: "OTHER", Value: []any{float64(3)}}}},
			want:    defaultChecksMaxDays,
		},
		{
			name:    "таблица настроек пуста",
			storage: fakeStorage{},
			want:    defaultChecksMaxDays,
		},
		{
			name:    "БД недоступна",
			storage: fakeStorage{err: errors.New("connection refused")},
			want:    defaultChecksMaxDays,
		},
		{
			name: "мусор в значении",
			storage: fakeStorage{settings: []modelsI.SettingsEntity{
				{Key: keyChecksMaxDays, Value: []any{"неделя"}},
			}},
			want: defaultChecksMaxDays,
		},
		{
			name: "ноль не должен убирать кнопки молча",
			storage: fakeStorage{settings: []modelsI.SettingsEntity{
				{Key: keyChecksMaxDays, Value: []any{float64(0)}},
			}},
			want: defaultChecksMaxDays,
		},
		{
			name: "отрицательное значение",
			storage: fakeStorage{settings: []modelsI.SettingsEntity{
				{Key: keyChecksMaxDays, Value: []any{float64(-5)}},
			}},
			want: defaultChecksMaxDays,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checksMaxDays(context.Background(), tt.storage, log); got != tt.want {
				t.Errorf("checksMaxDays = %d, ожидалось %d", got, tt.want)
			}
		})
	}
}
