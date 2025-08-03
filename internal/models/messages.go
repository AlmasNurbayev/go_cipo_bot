package models

import "time"

type MessagesType struct {
	Created_at   time.Time           `json:"created_at"`
	Sending_at   time.Time           `json:"sending_at"`
	IsSending    bool                `json:"is_sending"`
	UserId       int64               `json:"user_id"`
	Transactions []TransactionEntity `json:"transactions"`
}
