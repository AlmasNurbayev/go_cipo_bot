package models

import "github.com/guregu/null/v5"

type UserEntity struct {
	Id                 int64       `json:"id" db:"id"`
	Telegram_id        string      `json:"telegram_id" db:"telegram_id"`
	Telegram_name      string      `json:"telegram_name" db:"telegram_name"`
	Transaction_cursor null.String `json:"transaction_cursor" db:"transaction_cursor"`
}
