package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (s *Storage) ListKassa(ctx context.Context) ([]models.KassaEntity, error) {
	op := "storage.ListKassa"
	log := s.log.With(slog.String("op", op))

	query := `SELECT * FROM kassa;`

	var kassa []models.KassaEntity
	var err error
	// если есть транзакция, используем ее, иначе стандартный пул
	if s.Tx != nil {
		db := *s.Tx
		err = pgxscan.Select(ctx, db, &kassa, query)
	} else {
		db := s.Db
		err = pgxscan.Select(ctx, db, &kassa, query)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return kassa, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return kassa, err
	}
	return kassa, nil
}
