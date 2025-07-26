package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (s *Storage) ListUsers(ctx context.Context) ([]models.UserEntity, error) {
	op := "storage.ListUsers"
	log := s.log.With(slog.String("op", op))

	query := `SELECT * FROM users;`

	var users []models.UserEntity
	var err error
	// если есть транзакция, используем ее, иначе стандартный пул
	if s.Tx != nil {
		db := *s.Tx
		err = pgxscan.Select(ctx, db, &users, query)
	} else {
		db := s.Db
		err = pgxscan.Select(ctx, db, &users, query)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return users, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return users, err
	}
	return users, nil
}

func (s *Storage) SetCursor(ctx context.Context, cursorId int64, userId int64) error {
	op := "storage.SetCursor"
	log := s.log.With(slog.String("op", op))

	if userId == 0 || cursorId == 0 {
		return errors.New("user id or cursor id is empty")
	}
	query := `UPDATE users SET transaction_cursor = $1 WHERE id = $2;`

	// если есть транзакция, используем ее, иначе стандартный пул
	var err error
	if s.Tx != nil {
		db := *s.Tx
		_, err = db.Exec(ctx, query, cursorId, userId)
	} else {
		db := s.Db
		_, err = db.Exec(ctx, query, cursorId, userId)
	}
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return err
	}

	return nil
}
