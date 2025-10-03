package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (s *Storage) GetSettings(ctx context.Context) ([]models.SettingsEntity, error) {

	op := "storage.GetSettings"
	log := s.log.With(slog.String("op", op))

	query := `SELECT * FROM app_settings;`

	var settings []models.SettingsEntity

	db := s.Db
	err := pgxscan.Select(ctx, db, &settings, query)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return settings, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return settings, err
	}
	return settings, nil
}
