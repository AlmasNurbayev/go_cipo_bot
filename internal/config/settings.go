package config

import (
	"context"
	"log/slog"
	"reflect"

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

func GetSettings(storage *storage.Storage, log slog.Logger) ([]models.SettingsEntity, error) {
	var result []models.SettingsEntity
	settings, err := storage.GetSettings(context.Background())
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}
	return settings, nil
}

func GetOneKeySettings(key string, storage *storage.Storage, log slog.Logger) (models.SettingsEntity, error) {
	var result models.SettingsEntity
	settings, err := storage.GetSettings(context.Background())
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
