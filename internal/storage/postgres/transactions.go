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

func (s *Storage) InsertTransactions(ctx context.Context,
	transactions []models.TransactionEntity) (int, error) {

	op := "storage.InsertTransactions"
	log := s.log.With(slog.String("op", op))

	db := *s.Tx

	count := 0
	for _, transaction := range transactions {
		cmdTag, err := db.Exec(ctx, `
			INSERT INTO transactions (kassa_id, operationdate, sum_operation, type_operation,
				shift, subtype, systemdate, availablesum, offlinefiscalnumber,
				onlinefiscalnumber, paymenttypes, organization_id, ofd_id,
				ofd_name, knumber, cheque, images, cheque_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, $17, $18)
			ON CONFLICT (ofd_id) DO NOTHING
		`, transaction.Kassa_id, transaction.Operationdate,
			transaction.Sum_operation, transaction.Type_operation,
			transaction.Shift, transaction.Subtype, transaction.Systemdate,
			transaction.Availablesum, transaction.Offlinefiscalnumber,
			transaction.Onlinefiscalnumber, transaction.Paymenttypes,
			transaction.Organization_id,
			transaction.Ofd_id, transaction.Ofd_name, transaction.Knumber,
			transaction.Cheque, transaction.Images,
			transaction.ChequeJSON,
		)

		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return 0, err
		}
		count = count + int(cmdTag.RowsAffected())

	}
	log.Info("Inserted rows", slog.Int("rowsInserted", int(count)))

	return count, nil
}

func (s *Storage) ListTransactionsByDate(ctx context.Context,
	start time.Time, end time.Time) ([]models.TransactionEntity, error) {

	op := "storage.ListTransactionsByDate"
	log := s.log.With(slog.String("op", op))
	var transactions []models.TransactionEntity

	query := `SELECT * FROM transactions
	WHERE operationdate >= $1 AND operationdate < $2;`

	db := s.Db
	err := pgxscan.Select(ctx, db, &transactions, query, start, end)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return transactions, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return transactions, err
	}

	return transactions, nil
}

func (s *Storage) CheckExistsTransactions(ctx context.Context,
	kofd_ids []string) ([]string, error) {

	op := "storage.CheckExistsTransactions"
	log := s.log.With(slog.String("op", op))
	var existsIds []string

	query := `SELECT ofd_id FROM transactions
	WHERE ofd_id = ANY($1);`

	db := s.Db
	err := pgxscan.Select(ctx, db, &existsIds, query, kofd_ids)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// если выкидывается ошибка нет строк, возвращаем пустой массив
			return existsIds, nil
		}
		log.Error("error: ", slog.String("err", err.Error()))
		return existsIds, err
	}

	return existsIds, nil
}
