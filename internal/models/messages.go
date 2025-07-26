package models

import "time"

type MessagesType struct {
	Created_at   time.Time
	Sending_at   time.Time
	IsSending    bool
	UserId       int64
	Transactions []TransactionEntity
}
