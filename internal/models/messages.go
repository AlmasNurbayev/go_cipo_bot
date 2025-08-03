package models

import "time"

type MessagesType struct {
	Created_at   time.Time           `json:"created_at"`
	Sending_at   time.Time           `json:"sending_at"`
	UserId       int64               `json:"user_id"`
	Telegram_id  string              `json:"telegram_id"`
	Transactions []TransactionEntity `json:"transactions"`
}
