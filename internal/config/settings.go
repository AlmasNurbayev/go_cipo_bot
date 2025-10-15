package config

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	storage "github.com/AlmasNurbayev/go_cipo_bot/internal/storage/postgres"
)

func GetSettingsString(key string, settings []models.SettingsEntity) []string {
	for _, s := range settings {
		if s.Key == key {
			var result []string
			for _, v := range s.Value {
				switch val := v.(type) {
				case string:
					result = append(result, val)
				}
			}
			return result
		}
	}
	return nil
}

func GetSettingsFloat64(key string, settings []models.SettingsEntity) (float64, error) {
	for _, s := range settings {
		if s.Key == key {
			for _, v := range s.Value {
				switch val := v.(type) {
				case float64:
					return val, nil
				}
			}
		}
	}
	return 0, errors.New("no float64 value found for key: " + key)
}

func GetSettingsUSDRates(key string, settings []models.SettingsEntity) ([]models.USDRates, error) {
	var result []models.USDRates
	for _, s := range settings {
		if s.Key == key {
			for _, v := range s.Value {
				m, ok := v.(map[string]any)
				if !ok {
					continue
				}
				for year, rateVal := range m {
					rateFloat, ok := rateVal.(float64)
					if !ok {
						continue
					}
					year, err := strconv.Atoi(year)
					if err != nil {
						return result, err
					}
					result = append(result, models.USDRates{
						Year: year,
						Rate: int(rateFloat),
					})
				}
			}

		}
	}
	return result, nil
}

func GetSettingsGSheetsSources(key string, settings []models.SettingsEntity) []models.Books {
	var result []models.Books
	for _, s := range settings {
		if s.Key == key {
			for _, v := range s.Value {
				var b models.Books
				if reflect.TypeOf(v).Kind() == reflect.Map {
					m := v.(map[string]any)
					if book, ok := m["book"].(string); ok {
						b.Book = book
					}
					if sheet, ok := m["sheet"].(string); ok {
						b.Sheet = sheet
					}
					if rg, ok := m["range"].(string); ok {
						b.Range = rg
					}
				}
				result = append(result, b)
			}
		}
	}
	return result
}

func GetSettings(ctx context.Context, storage *storage.Storage, log slog.Logger) ([]models.SettingsEntity, error) {
	var result []models.SettingsEntity
	settings, err := storage.GetSettings(ctx)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	return settings, nil
}

func GetOneKeySettings(ctx context.Context, key string, storage *storage.Storage, log slog.Logger) (models.SettingsEntity, error) {
	var result models.SettingsEntity
	settings, err := storage.GetSettings(ctx)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	for _, s := range settings {
		if s.Key == key {
			result = s
		}
	}
	return result, nil
}
