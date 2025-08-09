package storage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Ctx context.Context // экспортная для создания транзакций снаружи
	Db  *pgxpool.Pool   // экспортная для создания транзакций снаружи
	log *slog.Logger
	Tx  *pgx.Tx
}

func NewStorage(ctx context.Context, DSN string, log1 *slog.Logger) (*Storage, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel()
	const op = "postgres.NewStorage"
	var shortDSN = ""
	if idx := strings.Index(DSN, "@"); idx != -1 {
		shortDSN = DSN[idx+1:]
	}

	log := log1.With(slog.String("op", op))
	log.Info("init storage " + shortDSN)

	newConfig, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		log.Error("not parse postgres DSN", slog.String("err", err.Error()))
		return nil, err
	}
	newConfig.MaxConns = 20

	db, err := pgxpool.NewWithConfig(ctx, newConfig)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// создание транзакции отключено, так как создается по необходимости в парсере
	// tx, err := db.Begin(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("%s: %w", op, err)
	// }

	return &Storage{Ctx: ctx, Db: db, log: log}, nil
}

func (s *Storage) Close() {
	const op = "postgres.Close"
	log := s.log.With(slog.String("op", op))

	if s.Db != nil {
		//s.log.Info("active Postgres conns", slog.Any("acquired_conns", s.db.Stats().OpenConnections))
		//s.tx.Rollback(context.Background())
		s.Db.Close()
		log.Warn("DB connection closed")
	}
}
