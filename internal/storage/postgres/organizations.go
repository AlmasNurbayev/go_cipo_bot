package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (s *Storage) ListOrganizations(ctx context.Context) ([]models.OrganizationEntity, error) {
	op := "storage.ListOrganizations"
	log := s.log.With(slog.String("op", op))

	query := `SELECT * FROM organizations;`

	var organizations []models.OrganizationEntity
	var err error
	// если есть транзакция, используем ее, иначе стандартный пул

	db := s.Db
	err = pgxscan.Select(ctx, db, &organizations, query)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return organizations, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return organizations, err
	}
	return organizations, nil
}
