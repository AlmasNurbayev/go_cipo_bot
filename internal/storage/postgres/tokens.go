package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (s *Storage) InsertToken(ctx context.Context, token string,
	bin string, exp int64, nbf int64) error {

	op := "storage.InsertToken"
	log := s.log.With(slog.String("op", op))

	query := `INSERT INTO tokens (token, bin, exp, nbf) VALUES ($1, $2, $3, $4);`

	_, err := s.Db.Exec(ctx, query, token, bin, exp, nbf)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return err
	}

	return nil
}

func (s *Storage) ListActiveTokens(ctx context.Context,
	bin string) ([]models.TokenEntity, error) {

	op := "storage.InsertToken"
	log := s.log.With(slog.String("op", op))

	query := `SELECT * FROM tokens WHERE bin = $1 AND exp > $2 ORDER BY created_at DESC;`

	var tokens []models.TokenEntity

	db := s.Db
	err := pgxscan.Select(ctx, db, &tokens, query, bin, time.Now().Unix()+10)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return tokens, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return tokens, err
	}
	return tokens, nil
}

func (s *Storage) DeleteOldTokens(ctx context.Context) error {

	op := "storage.DeleteOldTokens"
	log := s.log.With(slog.String("op", op))

	query := `DELETE FROM tokens WHERE created_at < NOW() - INTERVAL '2 hour';`

	_, err := s.Db.Exec(ctx, query)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return err
	}
	return nil
}
