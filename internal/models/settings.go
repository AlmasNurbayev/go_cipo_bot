package models

import (
	"time"

	"github.com/guregu/null/v5"
)

type SettingsEntity struct {
	Key       string      `json:"key" db:"key"`
	Value     []any       `json:"value" db:"value"`
	Caption   null.String `json:"caption" db:"caption"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

type Books struct {
	Book  string `json:"book"`
	Sheet string `json:"sheet"`
	Range string `json:"range"`
}

type USDRates struct {
	Year int `json:"year"`
	Rate int `json:"rate"`
}
