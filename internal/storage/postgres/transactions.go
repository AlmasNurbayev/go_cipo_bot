package storage

import (
	"context"
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func (s *Storage) InsertTransactions(ctx context.Context,
	transactions []models.TransactionEntity) error {

	op := "storage.InsertTransactions"
	log := s.log.With(slog.String("op", op))

	db := *s.Tx

	count := 0
	for _, transaction := range transactions {
		cmdTag, err := db.Exec(ctx, `
			INSERT INTO transactions (kassa_id, operationdate, sum_operation, type_operation,
				shift, subtype, systemdate, availablesum, offlinefiscalnumber,
				onlinefiscalnumber, paymenttypes, organization_id, ofd_id,
				ofd_name, knumber, cheque, images, names)
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
			transaction.Cheque, transaction.Images, transaction.Names)

		if err != nil {
			log.Error("error: ", slog.String("err", err.Error()))
			return err
		}
		count = count + int(cmdTag.RowsAffected())

	}
	log.Info("Inserted rows", slog.Int("rowsInserted", int(count)))

	return nil
}
