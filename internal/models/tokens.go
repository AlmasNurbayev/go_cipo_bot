package models

import (
	"time"

	"github.com/guregu/null/v5"
)

type TokenEntity struct {
	Id         int64     `json:"id" db:"id"`
	Bin        string    `json:"bin" db:"bin"`
	Token      string    `json:"token" db:"token"`
	Exp        int64     `json:"exp" db:"exp"`
	Nbf        int64     `json:"nbf" db:"nbf"`
	Working    null.Bool `json:"working" db:"working"`
	Created_at time.Time `json:"created_at" db:"created_at"`
}
